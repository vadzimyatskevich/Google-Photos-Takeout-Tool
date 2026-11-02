package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	_ "time/tzdata"
)

var cliMode bool

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ";") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	peek, explicit := peekLang(os.Args[1:])
	L := NormalizeLang(peek)
	o := defaultOptions()
	var srcs multiFlag
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", L.T("usage.header"))
		flag.PrintDefaults()
	}
	flag.Var(&srcs, "src", L.T("flag.src"))
	flag.StringVar(&o.Dest, "dest", "", L.T("flag.dest"))
	flag.StringVar(&o.Layout, "layout", o.Layout, L.T("flag.layout"))
	flag.StringVar(&o.Albums, "albums", o.Albums, L.T("flag.albums"))
	flag.BoolVar(&o.Rename, "rename", false, L.T("flag.rename"))
	flag.StringVar(&o.TZ, "tz", o.TZ, L.T("flag.tz"))
	flag.BoolVar(&o.Overwrite, "overwrite", false, L.T("flag.overwrite"))
	flag.BoolVar(&o.SkipEdited, "skip-edited", false, L.T("flag.skip-edited"))
	flag.BoolVar(&o.XMPSidecar, "xmp", o.XMPSidecar, L.T("flag.xmp"))
	flag.BoolVar(&o.Rating, "rating", o.Rating, L.T("flag.rating"))
	flag.BoolVar(&o.AlbumKeywords, "album-keywords", o.AlbumKeywords, L.T("flag.album-keywords"))
	flag.BoolVar(&o.People, "people", o.People, L.T("flag.people"))
	flag.BoolVar(&o.DryRun, "dry-run", false, L.T("flag.dry-run"))
	flag.BoolVar(&o.Force, "force", false, L.T("flag.force"))
	flag.IntVar(&o.Workers, "workers", 0, L.T("flag.workers"))
	flag.StringVar(&o.ExifTool, "exiftool", "", L.T("flag.exiftool"))
	flag.StringVar(&o.Lang, "lang", "en", L.T("flag.lang"))
	port := flag.Int("port", 0, L.T("flag.port"))
	noBrowser := flag.Bool("no-browser", false, L.T("flag.no-browser"))
	webUI := flag.Bool("web", false, L.T("flag.web"))
	flag.Parse()
	L = NormalizeLang(o.Lang)
	o.Lang = string(L)
	o.Sources = append(srcs, flag.Args()...)
	if o.Dest != "" {
		cliMode = true
		ensureConsole()
		p := &Progress{}
		err := Run(o, p)
		snap := p.snapshot()
		if s := p.Stats; s != nil {
			fmt.Printf(L.T("cli.summary"),
				s.Unique, s.Duplicates, s.Copied, s.Resumed, s.Errors, s.DateJSON, s.DateEXIF, s.DateName, s.DateYear, s.DateNone, s.GPSWritten)
		}
		fmt.Println(L.T("cli.report", snap["report"]))
		if err != nil {
			fmt.Fprintln(os.Stderr, L.T("cli.error", err))
			os.Exit(1)
		}
		return
	}
	if *webUI || !guiEnabled {
		ensureConsole()
		serveUI(*port, !*noBrowser, L, explicit)
		return
	}
	launchGUI(L, explicit)
}
