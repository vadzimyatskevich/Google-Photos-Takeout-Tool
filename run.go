package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	MediaEntries  int
	Unique        int
	Duplicates    int
	Incomplete    int
	EditedSkipped int
	DateJSON      int
	DateEXIF      int
	DateName      int
	DateYear      int
	DateNone      int
	GPSWritten    int
	DescWritten   int
	MetaWritten   int
	XMPSidecars   int
	Copied        int
	Resumed       int
	AlbumLinks    int
	AlbumCopies   int
	Trash         int
	Errors        int
	JSONTotal     int
	JSONUnmatched int
	MediaNoJSON   int
	UniqueBytes   int64
	NeedBytes     int64
	FreeBytes     int64
}

type Progress struct {
	mu         sync.Mutex
	Phase      string `json:"phase"`
	Done       int64  `json:"done"`
	Total      int64  `json:"total"`
	BytesDone  int64  `json:"bytesDone"`
	BytesTotal int64  `json:"bytesTotal"`
	Current    string `json:"current"`
	ETA        string `json:"eta"`
	Running    bool   `json:"running"`
	Finished   bool   `json:"finished"`
	DryRun     bool   `json:"dryRun"`
	Err        string `json:"error"`
	Warning    string `json:"warning"`
	Report     string `json:"report"`
	Log        string `json:"log"`
	Dest       string `json:"dest"`
	Stats      *Stats `json:"stats"`
	started    time.Time
	cancel     atomic.Bool
}

func (p *Progress) set(f func(p *Progress)) { p.mu.Lock(); f(p); p.mu.Unlock() }

func (p *Progress) snapshot() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	b, _ := json.Marshal(p)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

type stateRec struct {
	Path string `json:"path"`
	Date string `json:"date,omitempty"`
	Src  string `json:"src,omitempty"`
}

type Runner struct {
	o          Options
	loc        *time.Location
	p          *Progress
	st         Stats
	smu        sync.Mutex
	state      map[string]stateRec
	stMu       sync.Mutex
	stDirty    int
	reserved   map[string]bool
	resMu      sync.Mutex
	outDir     string // _restore dir (or temp dir for dry-run)
	logF       *os.File
	logMu      sync.Mutex
	etPath     string
	groups     []*Group
	unmatched  []*Sidecar
	noJSON     []*Group
	incomplete []*Entry
}

func (r *Runner) logf(format string, a ...any) {
	line := time.Now().Format("15:04:05 ") + fmt.Sprintf(format, a...)
	r.logMu.Lock()
	if r.logF != nil {
		fmt.Fprintln(r.logF, line)
	}
	r.logMu.Unlock()
	if cliMode {
		fmt.Println(line)
	}
}

func (r *Runner) stat(f func(s *Stats)) { r.smu.Lock(); f(&r.st); r.smu.Unlock() }

var errCancelled = errors.New("остановлено пользователем")

func Run(o Options, p *Progress) (err error) {
	r := &Runner{o: o, p: p, state: map[string]stateRec{}, reserved: map[string]bool{}}
	p.set(func(p *Progress) {
		p.Running, p.Finished, p.Err, p.Warning, p.DryRun, p.Phase, p.Done, p.Total, p.BytesDone, p.BytesTotal = true, false, "", "", o.DryRun, "Подготовка", 0, 0, 0, 0
		p.started = time.Now()
		p.Stats = nil
		p.Dest = o.Dest
	})
	p.cancel.Store(false)
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("внутренняя ошибка: %v", rec)
		}
		st := r.st
		p.set(func(p *Progress) {
			p.Running, p.Finished = false, true
			if err != nil {
				p.Err = err.Error()
				p.Phase = "Ошибка"
			} else {
				p.Phase = "Готово"
			}
			p.Stats = &st
			p.Current = ""
		})
		if r.logF != nil {
			if err != nil {
				r.logf("ОШИБКА: %v", err)
			}
			r.logF.Close()
		}
	}()
	if o.Dest == "" {
		return errors.New("не указана папка назначения")
	}
	if len(o.Sources) == 0 {
		return errors.New("не указаны источники")
	}
	o.Dest = filepath.Clean(o.Dest)
	r.o = o
	if r.loc, err = parseTZ(o.TZ); err != nil {
		return fmt.Errorf("часовой пояс %q: %v", o.TZ, err)
	}
	for _, s := range o.Sources {
		s = strings.Trim(strings.TrimSpace(s), "\"")
		if s == "" {
			continue
		}
		if a, b := absLower(s), absLower(o.Dest); a == b || strings.HasPrefix(b, a+string(filepath.Separator)) {
			return fmt.Errorf("папка назначения не может быть внутри источника: %s", s)
		}
	}
	if o.DryRun {
		r.outDir = filepath.Join(os.TempDir(), "TakeoutRestore-dryrun-"+time.Now().Format("20060102-150405"))
	} else {
		r.outDir = filepath.Join(o.Dest, "_restore")
	}
	if err := os.MkdirAll(r.outDir, 0o755); err != nil {
		return err
	}
	r.logF, _ = os.OpenFile(filepath.Join(r.outDir, "log.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	p.set(func(p *Progress) { p.Log = filepath.Join(r.outDir, "log.txt") })
	r.logf("=== Запуск (dry-run=%v) источники=%q назначение=%q структура=%s альбомы=%s TZ=%s", o.DryRun, o.Sources, o.Dest, o.Layout, o.Albums, r.loc)
	if b, e := os.ReadFile(filepath.Join(o.Dest, "_restore", "state.json")); e == nil {
		json.Unmarshal(b, &r.state)
		r.logf("Найдено состояние прошлого запуска: %d файлов", len(r.state))
	}
	if r.etPath, err = findExifTool(o.ExifTool); err != nil {
		return fmt.Errorf("exiftool не найден: %v", err)
	}
	r.logf("exiftool: %s", r.etPath)

	// 1. scan
	p.set(func(p *Progress) { p.Phase = "Сканирование источников" })
	sr := scanSources(o.Sources, func(s string) { p.set(func(p *Progress) { p.Current = s }) })
	defer func() {
		for _, z := range sr.Zips {
			z.Close()
		}
	}()
	for _, e := range sr.Errs {
		r.logf("ПРЕДУПРЕЖДЕНИЕ: %s", e)
	}
	r.logf("Найдено медиа: %d, json: %d, zip: %d", len(sr.Media), len(sr.JSON), len(sr.Zips))
	if len(sr.Media) == 0 {
		return errors.New("в источниках не найдено фото/видео")
	}

	// 2. CRC for plain files, detect incomplete extracted copies
	var plain []*Entry
	zipIdx := map[string][]*Entry{}
	for _, e := range sr.Media {
		if e.ZF == nil {
			plain = append(plain, e)
		} else {
			zipIdx[relKey(e)] = append(zipIdx[relKey(e)], e)
		}
	}
	if len(plain) > 0 {
		p.set(func(p *Progress) {
			p.Phase = "Проверка распакованных файлов"
			p.Total = int64(len(plain))
			p.Done = 0
		})
		var wg sync.WaitGroup
		ch := make(chan *Entry)
		for i := 0; i < defaultWorkers(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for e := range ch {
					c, err := fileCRC(e.FSPath)
					if err != nil {
						e.Incomplete = true
						r.logf("ОШИБКА чтения %s: %v", e.FSPath, err)
					}
					e.CRC = c
					p.set(func(p *Progress) { p.Done++; p.Current = e.FSPath })
				}
			}()
		}
		for _, e := range plain {
			ch <- e
		}
		close(ch)
		wg.Wait()
		for _, e := range plain {
			if zs, ok := zipIdx[relKey(e)]; ok {
				same := false
				for _, z := range zs {
					if z.Size == e.Size && z.CRC == e.CRC {
						same = true
					}
				}
				if !same {
					e.Incomplete = true
				}
			}
		}
	}
	var media []*Entry
	for _, e := range sr.Media {
		if e.Incomplete {
			r.incomplete = append(r.incomplete, e)
			continue
		}
		media = append(media, e)
	}

	// 3. sidecars
	p.set(func(p *Progress) {
		p.Phase = "Чтение JSON-метаданных"
		p.Total = int64(len(sr.JSON))
		p.Done = 0
	})
	mi := newMetaIndex()
	sort.SliceStable(sr.JSON, func(i, j int) bool { return sr.JSON[i].ZF != nil && sr.JSON[j].ZF == nil })
	for _, e := range sr.JSON {
		mi.addJSON(e)
		p.set(func(p *Progress) { p.Done++ })
	}
	var kept []*Entry
	editedSkipped := 0
	for _, e := range media {
		e.Edited = isEditedName(e.Name)
		if e.Edited && o.SkipEdited {
			editedSkipped++
			continue
		}
		e.SC, e.MatchHow = mi.Match(e.Dir, e.Name)
		if e.SC != nil {
			e.SC.Used = true
		}
		kept = append(kept, e)
	}
	usedTitles := map[string]bool{}
	for _, sc := range mi.All {
		if sc.Used {
			usedTitles[strings.ToLower(sc.Dir+"/"+sc.Title)] = true
		}
	}
	for _, sc := range mi.All {
		if !sc.Used && !usedTitles[strings.ToLower(sc.Dir+"/"+sc.Title)] {
			r.unmatched = append(r.unmatched, sc)
		}
	}
	r.groups = buildGroups(kept, mi)
	for _, g := range r.groups {
		if g.SC == nil {
			r.noJSON = append(r.noJSON, g)
		}
	}
	r.st.MediaEntries = len(sr.Media)
	r.st.Unique = len(r.groups)
	r.st.Incomplete = len(r.incomplete)
	r.st.EditedSkipped = editedSkipped
	r.st.Duplicates = len(kept) - len(r.groups)
	r.st.JSONTotal = len(mi.All)
	r.st.JSONUnmatched = len(r.unmatched)
	r.st.MediaNoJSON = len(r.noJSON)
	for _, g := range r.groups {
		r.st.UniqueBytes += g.Size
		if rec, ok := r.state[g.Key]; !ok || !exists(filepath.Join(o.Dest, rec.Path)) {
			r.st.NeedBytes += g.Size
			if o.Albums == "copy" {
				r.st.NeedBytes += g.Size * int64(len(g.Albums))
			}
		}
	}
	r.logf("Уникальных файлов: %d (%.2f ГБ), дубликатов: %d, недораспакованных копий пропущено: %d", r.st.Unique, gb(r.st.UniqueBytes), r.st.Duplicates, r.st.Incomplete)

	// 4. disk space
	if free, e := freeSpace(existingParent(o.Dest)); e == nil {
		r.st.FreeBytes = int64(free)
		if r.st.NeedBytes > int64(free)-200<<20 {
			w := fmt.Sprintf("Недостаточно места: нужно ≈%.1f ГБ, свободно %.1f ГБ", gb(r.st.NeedBytes), gb(int64(free)))
			r.logf("ПРЕДУПРЕЖДЕНИЕ: %s", w)
			p.set(func(p *Progress) { p.Warning = w })
			if !o.DryRun && !o.Force {
				r.writeReport()
				return errors.New(w + ". Освободите место или включите «всё равно продолжить».")
			}
		}
	}

	// 5. process
	err = r.process()
	r.saveState(true)
	r.writeReport()
	if !o.DryRun {
		os.RemoveAll(filepath.Join(r.outDir, "tmp"))
	}
	r.logf("=== Завершено. Скопировано %d, уже было %d, ошибок %d", r.st.Copied, r.st.Resumed, r.st.Errors)
	return err
}

func gb(b int64) float64 { return float64(b) / (1 << 30) }

func exists(p string) bool { _, err := os.Lstat(p); return err == nil }

func absLower(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		a = p
	}
	return strings.ToLower(filepath.Clean(a))
}

func existingParent(p string) string {
	p, _ = filepath.Abs(p)
	for {
		if exists(p) {
			return p
		}
		q := filepath.Dir(p)
		if q == p {
			return p
		}
		p = q
	}
}

func (r *Runner) process() error {
	o := r.o
	queues := map[string][]*Group{}
	var bytesTotal int64
	for _, g := range r.groups {
		c := g.Canon.Container
		queues[c] = append(queues[c], g)
		bytesTotal += g.Size
	}
	var qs [][]*Group
	for c, q := range queues {
		sort.Slice(q, func(i, j int) bool {
			if q[i].Canon.ZF != nil {
				return q[i].Canon.Order < q[j].Canon.Order
			}
			return q[i].Canon.FSPath < q[j].Canon.FSPath
		})
		if c == "" {
			for len(q) > 100 {
				qs = append(qs, q[:100])
				q = q[100:]
			}
		}
		qs = append(qs, q)
	}
	ch := make(chan []*Group, len(qs))
	for _, q := range qs {
		ch <- q
	}
	close(ch)
	phase := "Копирование и запись метаданных"
	if o.DryRun {
		phase = "Пробный прогон (ничего не записывается)"
	}
	start := time.Now()
	r.p.set(func(p *Progress) {
		p.Phase = phase
		p.Total = int64(len(r.groups))
		p.Done = 0
		p.BytesTotal = bytesTotal
		p.BytesDone = 0
	})
	workers := o.Workers
	if workers <= 0 {
		workers = defaultWorkers()
	}
	var wg sync.WaitGroup
	var firstErr error
	var emu sync.Mutex
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			et, err := startExifTool(r.etPath)
			if err != nil {
				emu.Lock()
				firstErr = err
				emu.Unlock()
				return
			}
			defer et.Close()
			for q := range ch {
				for _, g := range q {
					if r.p.cancel.Load() {
						return
					}
					r.p.set(func(p *Progress) { p.Current = g.Canon.Display() })
					r.processGroup(et, g)
					r.p.set(func(p *Progress) {
						p.Done++
						p.BytesDone += g.Size
						el := time.Since(start).Seconds()
						if p.BytesDone > 0 && el > 2 {
							rem := el * float64(p.BytesTotal-p.BytesDone) / float64(p.BytesDone)
							p.ETA = (time.Duration(rem) * time.Second).Round(time.Second).String()
						}
					})
				}
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	if r.p.cancel.Load() {
		return errCancelled
	}
	return nil
}

// alloc reserves a free path dir/name (adds _1, _2...). same(p) reports an existing file that is already ours.
func (r *Runner) alloc(dir, name string, same func(string) bool) (string, bool) {
	r.resMu.Lock()
	defer r.resMu.Unlock()
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 0; ; i++ {
		n := name
		if i > 0 {
			n = fmt.Sprintf("%s_%d%s", stem, i, ext)
		}
		p := filepath.Join(dir, n)
		k := strings.ToLower(p)
		if r.reserved[k] {
			continue
		}
		if exists(p) {
			if same != nil && same(p) {
				r.reserved[k] = true
				return p, true
			}
			continue
		}
		r.reserved[k] = true
		return p, false
	}
}

func (r *Runner) fail(g *Group, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	g.Err = strings.TrimSpace(strings.TrimSpace(g.Err + "; " + msg)[0:])
	g.Err = strings.TrimPrefix(g.Err, "; ")
	r.stat(func(s *Stats) { s.Errors++ })
	r.logf("ОШИБКА [%s]: %s", g.Canon.Display(), msg)
}

func (r *Runner) processGroup(et *ExifTool, g *Group) {
	o := r.o
	dest := o.Dest
	// resume
	if rec, ok := r.state[g.Key]; ok {
		full := filepath.Join(dest, rec.Path)
		if exists(full) {
			r.resMu.Lock()
			r.reserved[strings.ToLower(full)] = true
			r.resMu.Unlock()
			g.Path, g.Status, g.DateSrc = full, "уже было", rec.Src
			if t, err := time.Parse(time.RFC3339, rec.Date); err == nil {
				g.Date = t.In(r.loc)
			}
			r.countDate(g)
			r.stat(func(s *Stats) { s.Resumed++ })
			r.makeLinks(g)
			return
		}
	}
	var tmp string
	var tags map[string]string
	needRead := !o.DryRun || g.SC == nil || g.SC.Taken <= 0
	if needRead {
		var tdir string
		if o.DryRun {
			tdir, _ = os.MkdirTemp("", "trdry")
			defer os.RemoveAll(tdir)
		} else {
			tdir = filepath.Join(r.outDir, "tmp")
			os.MkdirAll(tdir, 0o755)
		}
		tmp = filepath.Join(tdir, g.Key+strings.ToLower(filepath.Ext(g.Name)))
		if err := copyEntry(g.Canon, tmp); err != nil {
			os.Remove(tmp)
			// try another copy of the same content
			okAlt := false
			for _, e := range g.Members {
				if e != g.Canon && copyEntry(e, tmp) == nil {
					okAlt = true
					break
				}
			}
			if !okAlt {
				g.Status = "ошибка"
				r.fail(g, "чтение: %v", err)
				return
			}
		}
		tags = et.ReadTags(tmp)
	}
	r.decideDate(g, tags)
	r.countDate(g)
	dir := filepath.Join(dest, mainDir(g, o.Layout))
	name := g.Name
	if o.Rename && (g.DateSrc == "json" || g.DateSrc == "exif" || g.DateSrc == "filename") {
		name = g.Date.Format("2006-01-02_150405_") + name
	}
	final, _ := r.alloc(dir, name, nil)
	g.Path = final
	if g.Trash {
		r.stat(func(s *Stats) { s.Trash++ })
	}
	if o.DryRun {
		g.Status = "будет скопирован"
		r.dryCountMeta(g, tags)
		r.stat(func(s *Stats) { s.Copied++ })
		r.makeLinks(g)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		g.Status = "ошибка"
		r.fail(g, "создание папки: %v", err)
		os.Remove(tmp)
		return
	}
	if err := os.Rename(tmp, final); err != nil {
		g.Status = "ошибка"
		r.fail(g, "перемещение: %v", err)
		os.Remove(tmp)
		return
	}
	g.Status = "скопирован"
	r.writeMeta(et, g, tags)
	if !g.Date.IsZero() {
		if err := setFileTimes(final, g.Date); err != nil {
			r.logf("не удалось установить время файла %s: %v", final, err)
		}
	}
	r.stat(func(s *Stats) { s.Copied++ })
	rel, _ := filepath.Rel(dest, final)
	rec := stateRec{Path: rel, Src: g.DateSrc}
	if !g.Date.IsZero() {
		rec.Date = g.Date.Format(time.RFC3339)
	}
	r.stMu.Lock()
	r.state[g.Key] = rec
	r.stDirty++
	r.stMu.Unlock()
	r.saveState(false)
	r.makeLinks(g)
}

func copyEntry(e *Entry, dst string) error {
	rc, err := e.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, rc)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n != e.Size {
		err = fmt.Errorf("размер %d вместо %d", n, e.Size)
	}
	return err
}

func (r *Runner) decideDate(g *Group, tags map[string]string) {
	if g.SC != nil && g.SC.Taken > 0 {
		g.Date, g.DateSrc = time.Unix(g.SC.Taken, 0).In(r.loc), "json"
		return
	}
	for _, k := range []string{"EXIF:DateTimeOriginal", "EXIF:CreateDate", "XMP:DateTimeOriginal", "XMP:CreateDate", "XMP:DateCreated", "QuickTime:CreationDate", "QuickTime:CreateDate", "QuickTime:MediaCreateDate"} {
		if v, ok := tags[k]; ok {
			if t, ok := parseExifDate(v, r.loc); ok {
				if strings.HasPrefix(k, "QuickTime:") && t.Year() < 1995 {
					continue
				}
				g.Date, g.DateSrc = t, "exif"
				return
			}
		}
	}
	for _, e := range g.Members {
		if t, ok := dateFromName(e.Name, r.loc); ok {
			g.Date, g.DateSrc = t, "filename"
			return
		}
	}
	if g.Year != 0 {
		g.Date, g.DateSrc = time.Date(g.Year, 1, 1, 12, 0, 0, 0, r.loc), "year"
		return
	}
	g.DateSrc = "none"
}

func (r *Runner) countDate(g *Group) {
	r.stat(func(s *Stats) {
		switch g.DateSrc {
		case "json":
			s.DateJSON++
		case "exif":
			s.DateEXIF++
		case "filename":
			s.DateName++
		case "year":
			s.DateYear++
		default:
			s.DateNone++
		}
	})
}

func (r *Runner) dryCountMeta(g *Group, tags map[string]string) {
	if g.SC != nil {
		r.stat(func(s *Stats) {
			if g.SC.HasGeo {
				s.GPSWritten++
				g.GPS = true
			}
			if g.SC.Description != "" {
				s.DescWritten++
			}
		})
	}
}

func (r *Runner) makeLinks(g *Group) {
	o := r.o
	if o.Albums == "none" || g.Path == "" || g.Status == "ошибка" {
		return
	}
	cst, err := os.Stat(g.Path)
	if err != nil && !o.DryRun {
		return
	}
	for i, a := range g.Albums {
		if o.Layout == "album" && i == 0 {
			continue
		}
		dir := filepath.Join(o.Dest, albumRoot(o.Layout), sanitize(a))
		if filepath.Dir(g.Path) == dir {
			continue
		}
		p, existing := r.alloc(dir, filepath.Base(g.Path), func(p string) bool {
			st, err := os.Stat(p)
			if err != nil || cst == nil {
				return false
			}
			return os.SameFile(st, cst) || (o.Albums == "copy" && st.Size() == cst.Size())
		})
		g.Links = append(g.Links, p)
		if existing || o.DryRun {
			if o.DryRun {
				r.stat(func(s *Stats) {
					if o.Albums == "copy" {
						s.AlbumCopies++
					} else {
						s.AlbumLinks++
					}
				})
			}
			continue
		}
		os.MkdirAll(dir, 0o755)
		if o.Albums == "hardlink" {
			if err := os.Link(g.Path, p); err == nil {
				r.stat(func(s *Stats) { s.AlbumLinks++ })
				continue
			} else {
				r.logf("жёсткая ссылка не удалась (%v), копирую: %s", err, p)
			}
		}
		if err := copyFile(g.Path, p); err != nil {
			r.fail(g, "альбом %s: %v", a, err)
			continue
		}
		if !g.Date.IsZero() {
			setFileTimes(p, g.Date)
		}
		r.stat(func(s *Stats) { s.AlbumCopies++ })
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	return out.Close()
}

func (r *Runner) saveState(force bool) {
	if r.o.DryRun {
		return
	}
	r.stMu.Lock()
	defer r.stMu.Unlock()
	if !force && r.stDirty < 25 {
		return
	}
	r.stDirty = 0
	b, _ := json.MarshalIndent(r.state, "", " ")
	tmp := filepath.Join(r.outDir, "state.json.tmp")
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, filepath.Join(r.outDir, "state.json"))
	}
}
