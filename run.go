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

type PreviewItem struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

const maxRecentPreview = 12

type Progress struct {
	mu          sync.Mutex
	Phase       string        `json:"phase"`
	Done        int64         `json:"done"`
	Total       int64         `json:"total"`
	BytesDone   int64         `json:"bytesDone"`
	BytesTotal  int64         `json:"bytesTotal"`
	FileBytes   int64         `json:"fileBytes"`
	FileSize    int64         `json:"fileSize"`
	Current     string        `json:"current"`
	CurrentNote string        `json:"currentNote,omitempty"`
	ETA         string        `json:"eta"`
	Running     bool          `json:"running"`
	Finished    bool          `json:"finished"`
	DryRun      bool          `json:"dryRun"`
	Err         string        `json:"error"`
	Warning     string        `json:"warning"`
	Report      string        `json:"report"`
	Log         string        `json:"log"`
	Dest        string        `json:"dest"`
	Lang        string        `json:"lang,omitempty"`
	Stopped     bool          `json:"stopped"`
	Paused      bool          `json:"paused"`
	Stats       *Stats        `json:"stats"`
	Recent      []PreviewItem `json:"recent,omitempty"`
	started     time.Time
	phaseWork   string // last active phase (restored after Pause)
	cancel      atomic.Bool
	pause       atomic.Bool
}

func (p *Progress) set(f func(p *Progress)) { p.mu.Lock(); f(p); p.mu.Unlock() }

func previewableImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".jpe", ".png", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}

func (p *Progress) pushRecent(path string) {
	if p == nil || path == "" || !previewableImage(path) {
		return
	}
	item := PreviewItem{Path: path, Name: filepath.Base(path)}
	p.set(func(p *Progress) {
		out := make([]PreviewItem, 0, maxRecentPreview)
		out = append(out, item)
		for _, it := range p.Recent {
			if it.Path == path {
				continue
			}
			out = append(out, it)
			if len(out) >= maxRecentPreview {
				break
			}
		}
		p.Recent = out
	})
}

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
	L          Lang
	folderLang Lang
	folders    Folders
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

func (r *Runner) logf(format string, a ...any) { r.log(fmt.Sprintf(format, a...)) }

func (r *Runner) log(msg string) {
	line := time.Now().Format("15:04:05 ") + msg
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

var errCancelled = errors.New("cancelled")

func (r *Runner) cancelled() bool { return r.p != nil && r.p.cancel.Load() }

func (r *Runner) paused() bool { return r.p != nil && r.p.pause.Load() }

// setPhase updates the visible phase (ignored while paused; restored on Resume).
func (r *Runner) setPhase(phase string) {
	if r.p == nil {
		return
	}
	r.p.set(func(p *Progress) {
		p.phaseWork = phase
		if !p.pause.Load() {
			p.Phase = phase
		}
	})
}

func (r *Runner) setCurrentNote(note string) {
	if r.p == nil {
		return
	}
	r.p.set(func(p *Progress) { p.CurrentNote = note })
}

func (r *Runner) bumpBytes(delta, fileDone, fileSize int64) {
	if r.p == nil || delta == 0 && fileSize == 0 {
		return
	}
	r.p.set(func(p *Progress) {
		p.BytesDone += delta
		if fileSize > 0 {
			p.FileBytes, p.FileSize = fileDone, fileSize
		}
		el := time.Since(p.started).Seconds()
		if p.BytesDone > 0 && el > 2 && p.BytesTotal > p.BytesDone {
			rem := el * float64(p.BytesTotal-p.BytesDone) / float64(p.BytesDone)
			p.ETA = (time.Duration(rem) * time.Second).Round(time.Second).String()
		}
	})
}

func Run(o Options, p *Progress) (err error) {
	L := NormalizeLang(o.Lang)
	o.Lang = string(L)
	r := &Runner{o: o, L: L, p: p, state: map[string]stateRec{}, reserved: map[string]bool{}}
	p.set(func(p *Progress) {
		p.Running, p.Finished, p.Err, p.Warning, p.DryRun, p.Phase, p.Done, p.Total, p.BytesDone, p.BytesTotal = true, false, "", "", o.DryRun, L.T("phase.prep"), 0, 0, 0, 0
		p.FileBytes, p.FileSize = 0, 0
		p.CurrentNote = ""
		p.Stopped = false
		p.Paused = false
		p.phaseWork = L.T("phase.prep")
		p.started = time.Now()
		p.Stats = nil
		p.Recent = nil
		p.Dest = o.Dest
		p.Lang = string(L)
	})
	p.pause.Store(false)
	defer func() {
		if rec := recover(); rec != nil {
			err = r.L.E("err.internal", rec)
		}
		st := r.st
		stopped := errors.Is(err, errCancelled)
		p.pause.Store(false)
		p.set(func(p *Progress) {
			p.Running, p.Finished = false, true
			p.Paused = false
			p.FileBytes, p.FileSize = 0, 0
			p.CurrentNote = ""
			if stopped {
				p.Stopped = true
				p.Err = ""
				p.Phase = r.L.T("phase.stopped")
			} else if err != nil {
				p.Err = err.Error()
				p.Phase = r.L.T("phase.error")
			} else {
				p.Phase = r.L.T("phase.done")
			}
			p.Stats = &st
			p.Current = ""
		})
		if r.logF != nil {
			if stopped {
				r.log(r.L.T("log.stopped"))
			} else if err != nil {
				r.log(r.L.T("log.error", err))
			}
			r.logF.Close()
		}
	}()
	if o.Dest == "" {
		return r.L.E("err.nodest")
	}
	if len(o.Sources) == 0 {
		return r.L.E("err.nosrc")
	}
	o.Dest = filepath.Clean(o.Dest)
	r.o = o
	if r.loc, err = parseTZ(o.TZ); err != nil {
		return r.L.E("err.tz", o.TZ, err)
	}
	for _, s := range o.Sources {
		s = strings.Trim(strings.TrimSpace(s), "\"")
		if s == "" {
			continue
		}
		if a, b := absLower(s), absLower(o.Dest); a == b || strings.HasPrefix(b, a+string(filepath.Separator)) {
			return r.L.E("err.destinside", s)
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
	r.log(r.L.T("log.start", o.DryRun, o.Sources, o.Dest, o.Layout, o.Albums, r.loc, o.Lang))
	if stored, hasLang, files, legacy, ok := loadState(filepath.Join(o.Dest, "_restore", "state.json")); ok {
		r.state = files
		r.chooseFolderLang(stored, hasLang, legacy)
		r.log(r.L.T("log.state", len(files)))
	} else {
		r.chooseFolderLang("", false, false)
	}
	if r.etPath, err = findExifTool(o.ExifTool); err != nil {
		if len(embeddedExifTool) > 0 && o.ExifTool == "" {
			return r.L.E("err.exifembed", err)
		}
		return r.L.E("err.exiftool", err)
	}
	r.log(r.L.T("log.exiftool", r.etPath))

	// 1. scan
	r.setPhase(r.L.T("phase.scan"))
	sr := scanSources(o.Sources, func(s string) { p.set(func(p *Progress) { p.Current = s }) }, r.cancelled)
	defer func() {
		for _, z := range sr.Zips {
			z.Close()
		}
	}()
	if r.cancelled() || sr.Cancelled {
		return errCancelled
	}
	for _, e := range sr.Errs {
		r.log(r.L.T("log.warn", e))
	}
	r.log(r.L.T("log.found", len(sr.Media), len(sr.JSON), len(sr.Zips)))
	if len(sr.Media) == 0 {
		return r.L.E("err.nomedia")
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
		r.setPhase(r.L.T("phase.check"))
		p.set(func(p *Progress) {
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
					if r.cancelled() {
						continue
					}
					c, err := fileCRC(e.FSPath, r.cancelled, r.paused)
					if errors.Is(err, errCancelled) || r.cancelled() {
						continue
					}
					if err != nil {
						e.Incomplete = true
						r.log(r.L.T("log.readerr", e.FSPath, err))
					}
					e.CRC = c
					p.set(func(p *Progress) { p.Done++; p.Current = e.FSPath })
				}
			}()
		}
		for _, e := range plain {
			if r.cancelled() {
				break
			}
			ch <- e
		}
		close(ch)
		wg.Wait()
		if r.cancelled() {
			return errCancelled
		}
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
	r.setPhase(r.L.T("phase.json"))
	p.set(func(p *Progress) {
		p.Total = int64(len(sr.JSON))
		p.Done = 0
	})
	mi := newMetaIndex()
	sort.SliceStable(sr.JSON, func(i, j int) bool { return sr.JSON[i].ZF != nil && sr.JSON[j].ZF == nil })
	for _, e := range sr.JSON {
		if r.cancelled() {
			return errCancelled
		}
		for r.paused() && !r.cancelled() {
			time.Sleep(100 * time.Millisecond)
		}
		if r.cancelled() {
			return errCancelled
		}
		mi.addJSON(e)
		p.set(func(p *Progress) { p.Done++ })
	}
	if r.cancelled() {
		return errCancelled
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
	r.groups = buildGroups(kept, mi, r.folders)
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
	r.log(r.L.T("log.unique", r.st.Unique, gb(r.st.UniqueBytes), r.st.Duplicates, r.st.Incomplete))

	// 4. disk space
	if free, e := freeSpace(existingParent(o.Dest)); e == nil {
		r.st.FreeBytes = int64(free)
		if r.st.NeedBytes > int64(free)-200<<20 {
			w := r.L.T("err.diskshort", gb(r.st.NeedBytes), gb(int64(free)))
			r.log(r.L.T("log.warn", w))
			p.set(func(p *Progress) { p.Warning = w })
			if !o.DryRun && !o.Force {
				r.writeReport()
				return r.L.E("err.diskmore", w)
			}
		}
	}

	// 5. process
	if r.cancelled() {
		return errCancelled
	}
	err = r.process()
	r.saveState(true)
	r.writeReport()
	if !o.DryRun {
		os.RemoveAll(filepath.Join(r.outDir, "tmp"))
	}
	r.log(r.L.T("log.done", r.st.Copied, r.st.Resumed, r.st.Errors))
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
	phase := r.L.T("phase.copy")
	if o.DryRun {
		phase = r.L.T("phase.dry")
	}
	start := time.Now()
	r.setPhase(phase)
	r.p.set(func(p *Progress) {
		p.Total = int64(len(r.groups))
		p.Done = 0
		p.BytesTotal = bytesTotal
		p.BytesDone = 0
		p.FileBytes, p.FileSize = 0, 0
		p.CurrentNote = ""
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
			if r.cancelled() {
				return
			}
			et, err := startExifTool(r.etPath)
			if err != nil {
				if r.cancelled() {
					return
				}
				emu.Lock()
				if firstErr == nil {
					firstErr = r.L.E("err.exifstart", r.etPath, err)
				}
				emu.Unlock()
				return
			}
			defer et.Close()
			stopWatch := r.watchExif(et)
			defer stopWatch()
			for q := range ch {
				for _, g := range q {
					if r.cancelled() {
						return
					}
					for r.paused() && !r.cancelled() {
						time.Sleep(100 * time.Millisecond)
					}
					if r.cancelled() {
						return
					}
					r.p.set(func(p *Progress) {
						p.Current = g.Canon.Display()
						p.FileBytes, p.FileSize = 0, g.Size
						p.CurrentNote = ""
					})
					r.processGroup(et, g)
					if r.cancelled() {
						return
					}
					credit := g.byteCredit
					if credit > g.Size {
						credit = g.Size
					}
					r.p.set(func(p *Progress) {
						p.Done++
						p.BytesDone += g.Size - credit
						p.FileBytes, p.FileSize = 0, 0
						p.CurrentNote = ""
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
	if r.cancelled() {
		return errCancelled
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

// watchExif kills this worker's exiftool soon after Stop, so a -stay_open
// command blocked in readUntil returns instead of running to the end.
func (r *Runner) watchExif(et *ExifTool) func() {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if r.cancelled() {
					et.Kill()
					return
				}
			}
		}
	}()
	return func() { close(done) }
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

func (r *Runner) fail(g *Group, msg string) {
	msg = strings.TrimSpace(msg)
	g.Err = strings.TrimSpace(g.Err + "; " + msg)
	g.Err = strings.TrimPrefix(g.Err, "; ")
	r.stat(func(s *Stats) { s.Errors++ })
	r.log(r.L.T("log.errfile", g.Canon.Display(), msg))
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
			g.Path, g.Status, g.DateSrc = full, statusResumed, rec.Src
			if t, err := time.Parse(time.RFC3339, rec.Date); err == nil {
				g.Date = t.In(r.loc)
			}
			r.countDate(g)
			r.stat(func(s *Stats) { s.Resumed++ })
			r.p.pushRecent(full)
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
		r.setCurrentNote(r.L.T("progress.copy"))
		n, err := r.copyEntry(g.Canon, tmp)
		g.byteCredit += n
		if err != nil {
			os.Remove(tmp)
			if errors.Is(err, errCancelled) || r.cancelled() {
				return
			}
			// try another copy of the same content
			okAlt := false
			for _, e := range g.Members {
				if r.cancelled() {
					return
				}
				if e == g.Canon {
					continue
				}
				r.setCurrentNote(r.L.T("progress.copy"))
				n2, err2 := r.copyEntry(e, tmp)
				g.byteCredit += n2
				if err2 == nil {
					okAlt = true
					break
				}
			}
			if !okAlt {
				os.Remove(tmp)
				if r.cancelled() {
					return
				}
				g.Status = statusError
				r.fail(g, r.L.T("err.read", err))
				return
			}
		}
		if r.cancelled() {
			os.Remove(tmp)
			return
		}
		r.setCurrentNote(r.L.T("progress.meta.read"))
		tags = et.ReadTags(tmp)
		if r.cancelled() {
			os.Remove(tmp)
			return
		}
		r.setCurrentNote("")
	}
	r.decideDate(g, tags)
	r.countDate(g)
	dir := filepath.Join(dest, mainDir(g, o.Layout, r.folders))
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
		g.Status = statusWillCopy
		r.dryCountMeta(g, tags)
		r.stat(func(s *Stats) { s.Copied++ })
		r.makeLinks(g)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		g.Status = statusError
		r.fail(g, r.L.T("err.mkdir", err))
		os.Remove(tmp)
		return
	}
	if err := os.Rename(tmp, final); err != nil {
		g.Status = statusError
		r.fail(g, r.L.T("err.move", err))
		os.Remove(tmp)
		return
	}
	g.Status = statusCopied
	r.setCurrentNote(r.L.T("progress.meta.write"))
	r.writeMeta(et, g, tags)
	r.setCurrentNote("")
	// Stop before the file is recorded: drop the partial copy so a re-run
	// writes it once, instead of leaving a second name beside it.
	if r.cancelled() {
		os.Remove(final)
		os.Remove(final + ".xmp")
		g.Path, g.Status = "", ""
		return
	}
	if !g.Date.IsZero() {
		if err := setFileTimes(final, g.Date); err != nil {
			r.log(r.L.T("log.mtime", final, err))
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
	r.p.pushRecent(final)
	r.makeLinks(g)
}

// copyEntry copies one entry and reports progress. credited is how many bytes were
// already added to Progress.BytesDone (so the caller can avoid double-counting).
func (r *Runner) copyEntry(e *Entry, dst string) (credited int64, err error) {
	rc, err := e.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	f, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	var batch int64
	var lastReport time.Time
	flush := func(force bool) {
		if batch == 0 {
			return
		}
		if !force && time.Since(lastReport) < 200*time.Millisecond && batch < 512<<10 {
			return
		}
		b := batch
		batch = 0
		credited += b
		lastReport = time.Now()
		r.bumpBytes(b, credited, e.Size)
	}
	n, err := copyStream(f, rc, r.cancelled, r.paused, func(nr int64) {
		batch += nr
		flush(false)
	})
	flush(true)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if errors.Is(err, errCancelled) || r.cancelled() {
		return credited, errCancelled
	}
	if err == nil && n != e.Size {
		err = r.L.E("err.size", n, e.Size)
	}
	return credited, err
}

func copyStream(dst io.Writer, src io.Reader, stop, paused func() bool, onBytes func(int64)) (int64, error) {
	buf := make([]byte, 256<<10)
	var n int64
	for {
		if stop != nil && stop() {
			return n, errCancelled
		}
		for paused != nil && paused() {
			if stop != nil && stop() {
				return n, errCancelled
			}
			time.Sleep(100 * time.Millisecond)
		}
		if stop != nil && stop() {
			return n, errCancelled
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			n += int64(nw)
			if onBytes != nil && nw > 0 {
				onBytes(int64(nw))
			}
			if ew != nil {
				return n, ew
			}
		}
		if er != nil {
			if er == io.EOF {
				return n, nil
			}
			return n, er
		}
	}
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
	if o.Albums == "none" || g.Path == "" || g.Status == statusError {
		return
	}
	cst, err := os.Stat(g.Path)
	if err != nil && !o.DryRun {
		return
	}
	for i, a := range g.Albums {
		if r.cancelled() {
			return
		}
		if o.Layout == "album" && i == 0 {
			continue
		}
		dir := filepath.Join(o.Dest, albumRoot(o.Layout, r.folders), sanitize(a))
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
				r.log(r.L.T("log.hardlink", err, p))
			}
		}
		if err := r.copyFile(g.Path, p); err != nil {
			if errors.Is(err, errCancelled) || r.cancelled() {
				return
			}
			r.fail(g, r.L.T("err.album", a, err))
			continue
		}
		if !g.Date.IsZero() {
			setFileTimes(p, g.Date)
		}
		r.stat(func(s *Stats) { s.AlbumCopies++ })
	}
}

func (r *Runner) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err = copyStream(out, in, r.cancelled, r.paused, nil); err != nil {
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
	files := r.state
	if files == nil {
		files = map[string]stateRec{}
	}
	b, _ := json.MarshalIndent(diskState{Lang: string(r.folderLang), Files: files}, "", " ")
	tmp := filepath.Join(r.outDir, "state.json.tmp")
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, filepath.Join(r.outDir, "state.json"))
	}
}

// diskState is _restore/state.json. Lang is the folder-name language for this
// destination and is kept on later runs so a library is not split across two schemes.
// Files is the resume map. Older builds stored that map at the top level; loadState still reads it.
type diskState struct {
	Lang  string              `json:"lang,omitempty"`
	Files map[string]stateRec `json:"files,omitempty"`
}

func loadState(path string) (stored Lang, hasLang bool, files map[string]stateRec, legacy bool, ok bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false, nil, false, false
	}
	var probe map[string]json.RawMessage
	if json.Unmarshal(b, &probe) != nil {
		return "", false, nil, false, false
	}
	files = map[string]stateRec{}
	if _, has := probe["files"]; has {
		var ds diskState
		if json.Unmarshal(b, &ds) != nil {
			return "", false, nil, false, false
		}
		if ds.Files != nil {
			files = ds.Files
		}
		stored, hasLang = parseLang(ds.Lang)
		return stored, hasLang, files, false, true
	}
	if json.Unmarshal(b, &files) != nil || files == nil {
		files = map[string]stateRec{}
	}
	return "", false, files, true, true
}

// chooseFolderLang picks the naming scheme for this destination.
// A language saved in state.json wins. Otherwise existing folders are sniffed.
// A legacy state file with no language marker (Russian-only builds) stays Russian.
func (r *Runner) chooseFolderLang(stored Lang, hasLang, legacy bool) {
	switch {
	case hasLang:
		r.folderLang = stored
	default:
		if det, ok := detectFolderLang(r.o.Dest); ok {
			r.folderLang = det
		} else if legacy {
			r.folderLang = LangRU
		} else {
			r.folderLang = r.L
		}
	}
	r.folders = r.folderLang.Folders()
	if r.folderLang.Folders() != r.L.Folders() {
		if r.folderLang == LangRU {
			r.log(r.L.T("log.lock.ru"))
		} else {
			r.log(r.L.T("log.lock.en"))
		}
	}
}
