package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (r *Runner) writeReport() {
	s := r.st
	L := r.L
	if f, err := os.Create(filepath.Join(r.outDir, "files.csv")); err == nil {
		f.WriteString("\uFEFF")
		w := csv.NewWriter(f)
		w.Comma = ';'
		w.Write([]string{
			L.T("csv.status"), L.T("csv.file"), L.T("csv.date"), L.T("csv.datesrc"), L.T("csv.json"),
			L.T("csv.match"), L.T("csv.albums"), L.T("csv.copies"), L.T("csv.source"), L.T("csv.error"),
		})
		for _, g := range r.groups {
			d := ""
			if !g.Date.IsZero() {
				d = g.Date.Format("2006-01-02 15:04:05 -07:00")
			}
			js, how := "", ""
			if g.SC != nil {
				js = g.SC.JSONName
				for _, e := range g.Members {
					if e.SC == g.SC {
						how = e.MatchHow
						break
					}
				}
			}
			w.Write([]string{L.statusText(g.Status), rel(r.o.Dest, g.Path), d, L.dateSrcText(g.DateSrc), js, how, strings.Join(g.Albums, " | "), fmt.Sprint(len(g.Members)), g.Canon.Display(), g.Err})
		}
		w.Flush()
		f.Close()
	}
	var b strings.Builder
	title := L.T("rep.title")
	if r.o.DryRun {
		title += L.T("rep.dry")
	}
	htmlLang := string(L)
	if htmlLang == "" {
		htmlLang = "en"
	}
	fmt.Fprintf(&b, `<!doctype html><html lang="%s"><meta charset="utf-8"><title>%s</title><style>
body{font-family:Segoe UI,Arial,sans-serif;margin:24px;color:#222;max-width:1200px}h1{font-size:22px}h2{font-size:17px;margin-top:28px}
table{border-collapse:collapse}td,th{border:1px solid #ddd;padding:4px 8px;font-size:13px;text-align:left;vertical-align:top}th{background:#f3f3f3}
.n{text-align:right;font-variant-numeric:tabular-nums}details{margin:8px 0}summary{cursor:pointer;font-weight:600}.err{color:#b00}
</style><h1>%s</h1><p>%s<br>%s</p><table>`,
		htmlLang, html.EscapeString(title), html.EscapeString(title), time.Now().Format(L.T("rep.timefmt")),
		L.T("rep.meta", html.EscapeString(r.o.Dest), html.EscapeString(strings.Join(r.o.Sources, "; ")), html.EscapeString(r.loc.String())))
	row := func(k string, v any) { fmt.Fprintf(&b, "<tr><td>%s</td><td class=n>%v</td></tr>", k, v) }
	row(L.T("rep.media"), s.MediaEntries)
	row(L.T("rep.unique"), s.Unique)
	row(L.T("rep.ubytes"), L.T("rep.gb", gb(s.UniqueBytes)))
	row(L.T("rep.dups"), s.Duplicates)
	row(L.T("rep.incomplete"), s.Incomplete)
	row(L.T("rep.edited"), s.EditedSkipped)
	row(L.T("rep.copied"), s.Copied)
	row(L.T("rep.resumed"), s.Resumed)
	row(L.T("rep.datejson"), s.DateJSON)
	row(L.T("rep.dateexif"), s.DateEXIF)
	row(L.T("rep.datename"), s.DateName)
	row(L.T("rep.dateyear"), s.DateYear)
	row(L.T("rep.datenone"), s.DateNone)
	row(L.T("rep.metafiles"), s.MetaWritten)
	row(L.T("rep.gps"), s.GPSWritten)
	row(L.T("rep.desc"), s.DescWritten)
	row(L.T("rep.xmp"), s.XMPSidecars)
	row(L.T("rep.links"), s.AlbumLinks)
	row(L.T("rep.copies"), s.AlbumCopies)
	row(L.T("rep.trash", r.folders.Trash), s.Trash)
	row(L.T("rep.jsontotal"), s.JSONTotal)
	row(L.T("rep.unmatched"), s.JSONUnmatched)
	row(L.T("rep.nojson"), s.MediaNoJSON)
	row("<span class=err>"+html.EscapeString(L.T("rep.errors"))+"</span>", s.Errors)
	if s.FreeBytes > 0 {
		row(L.T("rep.space"), L.T("rep.gbpair", gb(s.NeedBytes), gb(s.FreeBytes)))
	}
	b.WriteString("</table><p>")
	b.WriteString(L.T("rep.footer"))
	b.WriteString("</p>")
	list := func(title string, rows [][]string, head []string) {
		fmt.Fprintf(&b, "<details%s><summary>%s (%d)</summary><table><tr>", map[bool]string{true: " open", false: ""}[len(rows) > 0 && len(rows) < 50], html.EscapeString(title), len(rows))
		for _, h := range head {
			fmt.Fprintf(&b, "<th>%s</th>", html.EscapeString(h))
		}
		b.WriteString("</tr>")
		for _, rw := range rows {
			b.WriteString("<tr>")
			for _, c := range rw {
				fmt.Fprintf(&b, "<td>%s</td>", html.EscapeString(c))
			}
			b.WriteString("</tr>")
		}
		b.WriteString("</table></details>")
	}
	var errs, noj, unm, inc [][]string
	for _, g := range r.groups {
		if g.Err != "" {
			errs = append(errs, []string{g.Canon.Display(), g.Err})
		}
	}
	for _, g := range r.noJSON {
		noj = append(noj, []string{g.Canon.Display(), L.dateSrcText(g.DateSrc), rel(r.o.Dest, g.Path)})
	}
	for _, sc := range r.unmatched {
		unm = append(unm, []string{sc.Dir + "/" + sc.JSONName, sc.Title})
	}
	sort.Slice(unm, func(i, j int) bool { return unm[i][0] < unm[j][0] })
	for _, e := range r.incomplete {
		inc = append(inc, []string{e.FSPath, fmt.Sprint(e.Size)})
	}
	list(L.T("rep.sec.errors"), errs, []string{L.T("rep.h.file"), L.T("rep.h.error")})
	list(L.T("rep.sec.nojson"), noj, []string{L.T("rep.h.file"), L.T("rep.h.datesrc"), L.T("rep.h.dest")})
	list(L.T("rep.sec.unmatched"), unm, []string{L.T("rep.h.json"), L.T("rep.h.title")})
	list(L.T("rep.sec.incomplete"), inc, []string{L.T("rep.h.file"), L.T("rep.h.size")})
	os.WriteFile(filepath.Join(r.outDir, "report.html"), []byte(b.String()), 0o644)
	rp := filepath.Join(r.outDir, "report.html")
	r.p.set(func(p *Progress) { p.Report = rp })
	r.log(L.T("log.report", rp))
}

func rel(base, p string) string {
	if p == "" {
		return ""
	}
	if r, err := filepath.Rel(base, p); err == nil {
		return r
	}
	return p
}
