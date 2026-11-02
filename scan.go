package main

import (
	"archive/zip"
	"encoding/binary"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Entry is one file found in a source (inside a zip or in a folder).
type Entry struct {
	Container string    // zip path or "" for plain files
	ZF        *zip.File // non-nil for zip entries
	FSPath    string    // for plain files
	Rel       string    // path with '/' separators (inside zip, or relative to source root)
	Name      string    // base name
	Dir       string    // parent directory name (album folder)
	Size      int64
	CRC       uint32
	Order     int64 // offset inside zip, for sequential reading

	SC         *Sidecar // matched sidecar
	MatchHow   string
	Edited     bool
	Incomplete bool
}

func (e *Entry) Open() (io.ReadCloser, error) {
	if e.ZF != nil {
		return e.ZF.Open()
	}
	return os.Open(e.FSPath)
}

func (e *Entry) Display() string {
	if e.ZF != nil {
		return e.Container + " :: " + e.Rel
	}
	return e.FSPath
}

// cp866 upper half (0x80-0xFF)
var cp866 = []rune("АБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдежзийклмноп░▒▓│┤╡╢╖╕╣║╗╝╜╛┐└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀рстуфхцчшщъыьэюяЁёЄєЇїЎў°∙·√№¤■\u00a0")

func decodeCP866(b string) string {
	var sb strings.Builder
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c < 0x80 {
			sb.WriteByte(c)
		} else {
			sb.WriteRune(cp866[c-0x80])
		}
	}
	return sb.String()
}

// zipName returns the correctly decoded entry name.
func zipName(f *zip.File) string {
	// Info-ZIP Unicode Path extra field (0x7075)
	ex := f.Extra
	for len(ex) >= 4 {
		id := binary.LittleEndian.Uint16(ex[0:2])
		sz := int(binary.LittleEndian.Uint16(ex[2:4]))
		if 4+sz > len(ex) {
			break
		}
		if id == 0x7075 && sz > 5 && ex[4] == 1 {
			u := string(ex[9 : 4+sz])
			if utf8.ValidString(u) {
				return u
			}
		}
		ex = ex[4+sz:]
	}
	if f.Flags&0x800 != 0 || utf8.ValidString(f.Name) {
		return f.Name
	}
	return decodeCP866(f.Name)
}

var mediaExt = map[string]bool{}
var writableExt = map[string]string{} // ext -> kind: img, gif, qt

func init() {
	for _, e := range strings.Fields("jpg jpeg jpe png tif tiff heic heif webp cr2 cr3 nef arw dng orf rw2 raf thm") {
		mediaExt["."+e] = true
		writableExt["."+e] = "img"
	}
	mediaExt[".gif"] = true
	writableExt[".gif"] = "gif"
	for _, e := range strings.Fields("mp4 m4v mov 3gp 3g2") {
		mediaExt["."+e] = true
		writableExt["."+e] = "qt"
	}
	for _, e := range strings.Fields("bmp avi mkv wmv mpg mpeg mts m2ts webm") {
		mediaExt["."+e] = true
	}
}

func isMedia(name string) bool { return mediaExt[strings.ToLower(filepath.Ext(name))] }

type ScanResult struct {
	Media []*Entry
	JSON  []*Entry
	Zips  []*zip.ReadCloser
	Errs  []string
}

func splitRel(rel string) (dir, name string) {
	rel = strings.TrimSuffix(rel, "/")
	i := strings.LastIndex(rel, "/")
	name = rel[i+1:]
	if i <= 0 {
		return "", name
	}
	d := rel[:i]
	j := strings.LastIndex(d, "/")
	return d[j+1:], name
}

func scanSources(srcs []string, prog func(string)) *ScanResult {
	r := &ScanResult{}
	add := func(e *Entry) {
		e.Dir, e.Name = splitRel(e.Rel)
		low := strings.ToLower(e.Name)
		if strings.HasSuffix(low, ".json") {
			r.JSON = append(r.JSON, e)
		} else if isMedia(low) {
			r.Media = append(r.Media, e)
		}
	}
	addZip := func(p string) {
		prog(p)
		zr, err := zip.OpenReader(p)
		if err != nil {
			r.Errs = append(r.Errs, "zip: "+p+": "+err.Error())
			return
		}
		r.Zips = append(r.Zips, zr)
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}
			off, _ := f.DataOffset()
			add(&Entry{Container: p, ZF: f, Rel: strings.ReplaceAll(zipName(f), "\\", "/"), Size: int64(f.UncompressedSize64), CRC: f.CRC32, Order: off})
		}
	}
	for _, s := range srcs {
		s = strings.Trim(strings.TrimSpace(s), "\"")
		if s == "" {
			continue
		}
		st, err := os.Stat(s)
		if err != nil {
			r.Errs = append(r.Errs, "source: "+s+": "+err.Error())
			continue
		}
		if !st.IsDir() {
			if strings.EqualFold(filepath.Ext(s), ".zip") {
				addZip(s)
			}
			continue
		}
		filepath.WalkDir(s, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				r.Errs = append(r.Errs, "walk: "+p+": "+err.Error())
				return nil
			}
			if d.IsDir() {
				if strings.EqualFold(d.Name(), "_restore") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(filepath.Ext(p), ".zip") {
				addZip(p)
				return nil
			}
			rel, _ := filepath.Rel(s, p)
			rel = filepath.ToSlash(rel)
			if !(isMedia(p) || strings.EqualFold(filepath.Ext(p), ".json")) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			add(&Entry{FSPath: p, Rel: rel, Size: info.Size()})
			return nil
		})
	}
	return r
}

// relKey: album/name lowercased for matching folder copies against zip copies.
func relKey(e *Entry) string { return strings.ToLower(e.Dir + "/" + e.Name) }

func fileCRC(p string) (uint32, error) {
	f, err := os.Open(p)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	h := crc32.NewIEEE()
	_, err = io.Copy(h, f)
	return h.Sum32(), err
}
