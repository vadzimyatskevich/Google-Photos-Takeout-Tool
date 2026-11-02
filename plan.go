package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	Sources       []string `json:"sources"`
	Dest          string   `json:"dest"`
	Layout        string   `json:"layout"` // ym | y_m | ymd | y | album
	Albums        string   `json:"albums"` // hardlink | copy | none
	Rename        bool     `json:"rename"`
	TZ            string   `json:"tz"`
	Overwrite     bool     `json:"overwrite"`
	SkipEdited    bool     `json:"skipEdited"`
	XMPSidecar    bool     `json:"xmpSidecar"`
	Rating        bool     `json:"rating"`
	AlbumKeywords bool     `json:"albumKeywords"`
	People        bool     `json:"people"`
	DryRun        bool     `json:"dryRun"`
	Force         bool     `json:"force"`
	Workers       int      `json:"workers"`
	Lang          string   `json:"lang"`
	ExifTool      string   `json:"-"`
}

func defaultOptions() Options {
	return Options{Layout: "ym", Albums: "hardlink", TZ: "local", XMPSidecar: true, Rating: true, AlbumKeywords: true, People: true, Lang: "en"}
}

type Group struct {
	Key     string
	Size    int64
	Members []*Entry
	Canon   *Entry
	SC      *Sidecar
	Name    string
	Albums  []string
	Trash   bool // only present in Trash
	Year    int  // from "Photos from YYYY"

	// results
	Path    string
	Date    time.Time
	DateSrc string // json | exif | filename | year | none
	Status  string
	Err     string
	GPS     bool
	Links   []string
}

func isTrashDir(d string) bool {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "trash", "корзина", "bin", "papierkorb", "corbeille", "papelera", "cestino", "prullenbak", "kosz":
		return true
	}
	return false
}

func isArchiveDir(d string) bool {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "archive", "архив", "archiv":
		return true
	}
	return false
}

var badChars = strings.NewReplacer("<", "_", ">", "_", ":", "_", "\"", "_", "/", "_", "\\", "_", "|", "_", "?", "_", "*", "_")

func sanitize(s string) string {
	s = badChars.Replace(s)
	s = strings.Map(func(r rune) rune {
		if r < 32 {
			return '_'
		}
		return r
	}, s)
	s = strings.TrimRight(strings.TrimSpace(s), ". ")
	if s == "" {
		s = "_"
	}
	up := strings.ToUpper(strings.TrimSuffix(s, filepath.Ext(s)))
	switch up {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "LPT1", "LPT2", "LPT3":
		s = "_" + s
	}
	if r := []rune(s); len(r) > 150 {
		ext := filepath.Ext(s)
		s = string(r[:150-len([]rune(ext))]) + ext
	}
	return s
}

func albumDisplay(dir string, mi *MetaIndex, f Folders) string {
	if t, ok := mi.AlbumTitles[strings.ToLower(dir)]; ok && strings.TrimSpace(t) != "" {
		return strings.TrimSpace(t)
	}
	if isArchiveDir(dir) {
		return f.Archive
	}
	return dir
}

func matchRank(how string) int {
	switch {
	case how == "name" || how == "title":
		return 0
	case how == "truncated":
		return 1
	case strings.HasPrefix(how, "edited"):
		return 2
	case how == "companion":
		return 3
	}
	return 9
}

func buildGroups(media []*Entry, mi *MetaIndex, f Folders) []*Group {
	byKey := map[string]*Group{}
	var order []*Group
	for _, e := range media {
		k := fmt.Sprintf("%d-%08x", e.Size, e.CRC)
		g := byKey[k]
		if g == nil {
			g = &Group{Key: k, Size: e.Size}
			byKey[k] = g
			order = append(order, g)
		}
		g.Members = append(g.Members, e)
	}
	for _, g := range order {
		bestRank := 99
		albums := map[string]bool{}
		trashOnly := true
		for _, e := range g.Members {
			if e.SC != nil && matchRank(e.MatchHow) < bestRank {
				g.SC, bestRank = e.SC, matchRank(e.MatchHow)
			}
			if y := yearFromDir(e.Dir); y != 0 {
				g.Year = y
				trashOnly = false
				continue
			}
			if isTrashDir(e.Dir) {
				continue
			}
			trashOnly = false
			if e.Dir != "" {
				a := albumDisplay(e.Dir, mi, f)
				if !albums[a] {
					albums[a] = true
					g.Albums = append(g.Albums, a)
				}
			}
		}
		g.Trash = trashOnly
		sort.Strings(g.Albums)
		// canonical: prefer plain files (fast), then zip order
		g.Canon = g.Members[0]
		for _, e := range g.Members {
			if e.ZF == nil && g.Canon.ZF != nil {
				g.Canon = e
			}
		}
		// name: shortest without "(n)"
		best := ""
		for _, e := range g.Members {
			n := e.Name
			_, num := stripNum(strings.TrimSuffix(n, filepath.Ext(n)))
			score := len(n)
			if num != "" {
				score += 1000
			}
			if best == "" || score < nameScore(best) {
				best = n
			}
		}
		// restore truncated filename from json title
		if g.SC != nil && bestRank <= 1 {
			t := g.SC.Title
			if strings.EqualFold(filepath.Ext(t), filepath.Ext(best)) && len(t) > len(best) &&
				strings.HasPrefix(strings.ToLower(t), strings.ToLower(strings.TrimSuffix(best, filepath.Ext(best)))) {
				best = t
			}
		}
		g.Name = sanitize(best)
	}
	return order
}

func nameScore(n string) int {
	_, num := stripNum(strings.TrimSuffix(n, filepath.Ext(n)))
	if num != "" {
		return len(n) + 1000
	}
	return len(n)
}

// mainDir returns the folder (relative to dest) for the group's physical copy.
func mainDir(g *Group, layout string, f Folders) string {
	if g.Trash {
		return f.Trash
	}
	t := g.Date
	switch g.DateSrc {
	case "none", "":
		if layout == "album" && len(g.Albums) > 0 {
			return sanitize(g.Albums[0])
		}
		return f.NoDate
	case "year":
		y := fmt.Sprintf("%04d", g.Year)
		if layout == "album" {
			if len(g.Albums) > 0 {
				return sanitize(g.Albums[0])
			}
			return y
		}
		if layout == "y" {
			return y
		}
		return filepath.Join(y, f.UnknownMonth)
	}
	y := t.Format("2006")
	switch layout {
	case "y_m":
		return filepath.Join(y, t.Format("01"))
	case "ymd":
		return filepath.Join(y, t.Format("2006-01-02"))
	case "y":
		return y
	case "album":
		if len(g.Albums) > 0 {
			return sanitize(g.Albums[0])
		}
		return y
	}
	return filepath.Join(y, t.Format("2006-01"))
}

func albumRoot(layout string, f Folders) string {
	if layout == "album" {
		return ""
	}
	return f.Albums
}
