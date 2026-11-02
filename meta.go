package main

import (
	"encoding/json"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Sidecar struct {
	JSONName    string
	Dir         string
	Title       string
	Description string
	Taken       int64
	Lat, Lon    float64
	Alt         float64
	HasGeo      bool
	Favorited   bool
	People      []string
	Used        bool
	Src         string
}

type tsField struct {
	Timestamp json.RawMessage `json:"timestamp"`
}

func (t tsField) val() int64 {
	s := strings.Trim(string(t.Timestamp), "\" ")
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

type geo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type rawJSON struct {
	Title          *string  `json:"title"`
	Description    string   `json:"description"`
	PhotoTakenTime *tsField `json:"photoTakenTime"`
	CreationTime   *tsField `json:"creationTime"`
	GeoData        *geo     `json:"geoData"`
	GeoDataExif    *geo     `json:"geoDataExif"`
	Favorited      bool     `json:"favorited"`
	People         []struct {
		Name string `json:"name"`
	} `json:"people"`
}

type DirIdx struct {
	byName  map[string]*Sidecar
	byTitle map[string][]*Sidecar
	list    []*Sidecar
}

type MetaIndex struct {
	dirs        map[string]*DirIdx
	AlbumTitles map[string]string // dir(lower) -> album title from metadata.json
	All         []*Sidecar
}

func newMetaIndex() *MetaIndex {
	return &MetaIndex{dirs: map[string]*DirIdx{}, AlbumTitles: map[string]string{}}
}

func (m *MetaIndex) addJSON(e *Entry) {
	rc, err := e.Open()
	if err != nil {
		return
	}
	data, _ := io.ReadAll(io.LimitReader(rc, 4<<20))
	rc.Close()
	var r rawJSON
	if json.Unmarshal(data, &r) != nil {
		return
	}
	dk := strings.ToLower(e.Dir)
	if r.PhotoTakenTime == nil && r.CreationTime == nil {
		if r.Title != nil && *r.Title != "" && e.Dir != "" {
			m.AlbumTitles[dk] = *r.Title
		}
		return
	}
	d := m.dirs[dk]
	if d == nil {
		d = &DirIdx{byName: map[string]*Sidecar{}, byTitle: map[string][]*Sidecar{}}
		m.dirs[dk] = d
	}
	ln := strings.ToLower(e.Name)
	if _, dup := d.byName[ln]; dup {
		return // same json from another export / zip part
	}
	sc := &Sidecar{JSONName: e.Name, Dir: e.Dir, Description: strings.TrimSpace(r.Description), Favorited: r.Favorited, Src: e.Display()}
	if r.Title != nil {
		sc.Title = *r.Title
	}
	if r.PhotoTakenTime != nil {
		sc.Taken = r.PhotoTakenTime.val()
	}
	if sc.Taken <= 0 && r.CreationTime != nil {
		sc.Taken = r.CreationTime.val()
	}
	for _, g := range []*geo{r.GeoDataExif, r.GeoData} {
		if g != nil && (g.Latitude != 0 || g.Longitude != 0) {
			sc.Lat, sc.Lon, sc.Alt, sc.HasGeo = g.Latitude, g.Longitude, g.Altitude, true
			break
		}
	}
	for _, p := range r.People {
		if p.Name != "" {
			sc.People = append(sc.People, p.Name)
		}
	}
	d.byName[ln] = sc
	if sc.Title != "" {
		lt := strings.ToLower(sc.Title)
		d.byTitle[lt] = append(d.byTitle[lt], sc)
	}
	d.list = append(d.list, sc)
	m.All = append(m.All, sc)
}

var numRe = regexp.MustCompile(`^(.*)(\(\d+\))$`)

func stripNum(s string) (string, string) {
	if m := numRe.FindStringSubmatch(s); m != nil {
		return m[1], m[2]
	}
	return s, ""
}

var editedSuffixes = []string{"-edited", "-изменено", "-отредактировано", "-bearbeitet", "-modifié", "-editado", "-modificato", "-bewerkt", "-redigeret", "-redigerad", "-muokattu", "-edytowane", "-upravené", "-upraveno", "-編集済み", "-편집됨", "-已修改", "-düzenlendi", "-змінено", "-redigert"}

// stripEdited returns name without edited suffix (on the lowercase stem).
func stripEdited(stem string) (string, bool) {
	b, num := stripNum(stem)
	for _, s := range editedSuffixes {
		if strings.HasSuffix(b, s) {
			return strings.TrimSuffix(b, s) + num, true
		}
	}
	return stem, false
}

func isEditedName(name string) bool {
	l := strings.ToLower(name)
	_, ed := stripEdited(strings.TrimSuffix(l, filepath.Ext(l)))
	return ed
}

// jsonStem: lowercase json filename without ".json", with trailing "(n)" split off.
func jsonStem(sc *Sidecar) (string, string) {
	s := strings.TrimSuffix(strings.ToLower(sc.JSONName), ".json")
	return stripNum(s)
}

func (m *MetaIndex) Match(dir, name string) (*Sidecar, string) {
	d := m.dirs[strings.ToLower(dir)]
	if d == nil {
		return nil, ""
	}
	n := strings.ToLower(name)
	if sc, how := d.match(n); sc != nil {
		return sc, how
	}
	ext := filepath.Ext(n)
	stem := strings.TrimSuffix(n, ext)
	if orig, ok := stripEdited(stem); ok {
		if sc, how := d.match(orig + ext); sc != nil {
			return sc, "edited→" + how
		}
	}
	// companion / live photo: same stem, other extension
	base, num := stripNum(stem)
	if o, ok := stripEdited(base); ok {
		base = o
	}
	var best *Sidecar
	for _, sc := range d.list {
		t := strings.ToLower(sc.Title)
		ts := strings.TrimSuffix(t, filepath.Ext(t))
		js, jn := jsonStem(sc)
		js = strings.TrimSuffix(js, ".supplemental-metadata")
		js = strings.TrimSuffix(js, filepath.Ext(js))
		if (ts == base && num == "") || (js == base && jn == num) {
			if best == nil || isPhotoExt(filepath.Ext(t)) {
				best = sc
			}
		}
	}
	if best != nil {
		return best, "companion"
	}
	return nil, ""
}

func isPhotoExt(e string) bool {
	k := writableExt[e]
	return k == "img" || k == "gif"
}

func (d *DirIdx) match(n string) (*Sidecar, string) {
	ext := filepath.Ext(n)
	stem := strings.TrimSuffix(n, ext)
	base, num := stripNum(stem)
	nn := base + ext
	cands := []string{n + ".json", n + ".supplemental-metadata.json"}
	if num != "" {
		cands = append(cands, nn+num+".json", nn+".supplemental-metadata"+num+".json", base+num+".json")
	}
	cands = append(cands, stem+".json")
	for _, c := range cands {
		if sc := d.byName[c]; sc != nil {
			return sc, "name"
		}
	}
	if num == "" {
		if l := d.byTitle[n]; len(l) > 0 {
			return l[0], "title"
		}
	}
	// truncated json names
	var best *Sidecar
	bestLen := 0
	fulls := []string{nn + ".supplemental-metadata", nn, base}
	for _, sc := range d.list {
		js, jn := jsonStem(sc)
		if jn != num || len(js) < 40 || len(js) <= bestLen {
			continue
		}
		for _, f := range fulls {
			if len(f) > len(js) && strings.HasPrefix(f, js) {
				best, bestLen = sc, len(js)
				break
			}
		}
	}
	if best != nil {
		return best, "truncated"
	}
	return nil, ""
}
