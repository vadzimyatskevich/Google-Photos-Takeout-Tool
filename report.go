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
	// CSV with every unique file
	if f, err := os.Create(filepath.Join(r.outDir, "files.csv")); err == nil {
		f.WriteString("\uFEFF")
		w := csv.NewWriter(f)
		w.Comma = ';'
		w.Write([]string{"статус", "файл", "дата", "источник даты", "json", "сопоставление json", "альбомы", "копий в источниках", "источник", "ошибка"})
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
			w.Write([]string{g.Status, rel(r.o.Dest, g.Path), d, g.DateSrc, js, how, strings.Join(g.Albums, " | "), fmt.Sprint(len(g.Members)), g.Canon.Display(), g.Err})
		}
		w.Flush()
		f.Close()
	}
	var b strings.Builder
	title := "Отчёт о восстановлении"
	if r.o.DryRun {
		title += " — ПРОБНЫЙ ПРОГОН (ничего не записано)"
	}
	fmt.Fprintf(&b, `<!doctype html><html lang="ru"><meta charset="utf-8"><title>%s</title><style>
body{font-family:Segoe UI,Arial,sans-serif;margin:24px;color:#222;max-width:1200px}h1{font-size:22px}h2{font-size:17px;margin-top:28px}
table{border-collapse:collapse}td,th{border:1px solid #ddd;padding:4px 8px;font-size:13px;text-align:left;vertical-align:top}th{background:#f3f3f3}
.n{text-align:right;font-variant-numeric:tabular-nums}details{margin:8px 0}summary{cursor:pointer;font-weight:600}.err{color:#b00}
</style><h1>%s</h1><p>%s<br>Назначение: <b>%s</b><br>Источники: %s<br>Часовой пояс: %s</p><table>`,
		title, title, time.Now().Format("02.01.2006 15:04"), html.EscapeString(r.o.Dest), html.EscapeString(strings.Join(r.o.Sources, "; ")), html.EscapeString(r.loc.String()))
	row := func(k string, v any) { fmt.Fprintf(&b, "<tr><td>%s</td><td class=n>%v</td></tr>", k, v) }
	row("Медиафайлов найдено в источниках (всех копий)", s.MediaEntries)
	row("Уникальных файлов", s.Unique)
	row("Объём уникальных файлов", fmt.Sprintf("%.2f ГБ", gb(s.UniqueBytes)))
	row("Дубликатов пропущено (тот же файл в другом альбоме/архиве/выгрузке)", s.Duplicates)
	row("Недораспакованных/битых копий в папках пропущено", s.Incomplete)
	row("Отредактированных версий пропущено", s.EditedSkipped)
	row("Скопировано", s.Copied)
	row("Уже было (повторный запуск)", s.Resumed)
	row("Дата из JSON (Google)", s.DateJSON)
	row("Дата из EXIF файла", s.DateEXIF)
	row("Дата из имени файла", s.DateName)
	row("Только год (папка «Photos from ГГГГ»)", s.DateYear)
	row("Без даты", s.DateNone)
	row("Записаны метаданные (файлов)", s.MetaWritten)
	row("Записан GPS", s.GPSWritten)
	row("Записано описание", s.DescWritten)
	row("XMP-файлов рядом (avi, bmp…)", s.XMPSidecars)
	row("Ссылок в альбомах (жёсткие ссылки)", s.AlbumLinks)
	row("Копий в альбомах", s.AlbumCopies)
	row("В «_Корзина»", s.Trash)
	row("JSON-метаданных", s.JSONTotal)
	row("JSON без файла", s.JSONUnmatched)
	row("Файлов без JSON", s.MediaNoJSON)
	row("<span class=err>Ошибок</span>", s.Errors)
	if s.FreeBytes > 0 {
		row("Нужно места / свободно", fmt.Sprintf("%.1f / %.1f ГБ", gb(s.NeedBytes), gb(s.FreeBytes)))
	}
	b.WriteString("</table><p>Полный список файлов: <code>files.csv</code> (открывается в Excel), журнал: <code>log.txt</code>.</p>")
	list := func(title string, rows [][]string, head []string) {
		fmt.Fprintf(&b, "<details%s><summary>%s (%d)</summary><table><tr>", map[bool]string{true: " open", false: ""}[len(rows) > 0 && len(rows) < 50], title, len(rows))
		for _, h := range head {
			fmt.Fprintf(&b, "<th>%s</th>", h)
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
		errs2 := g.DateSrc
		noj = append(noj, []string{g.Canon.Display(), errs2, rel(r.o.Dest, g.Path)})
	}
	for _, sc := range r.unmatched {
		unm = append(unm, []string{sc.Dir + "/" + sc.JSONName, sc.Title})
	}
	sort.Slice(unm, func(i, j int) bool { return unm[i][0] < unm[j][0] })
	for _, e := range r.incomplete {
		inc = append(inc, []string{e.FSPath, fmt.Sprint(e.Size)})
	}
	list("Ошибки", errs, []string{"файл", "ошибка"})
	list("Файлы без JSON", noj, []string{"файл", "источник даты", "куда"})
	list("JSON без медиафайла", unm, []string{"json", "title"})
	list("Пропущенные недораспакованные копии (есть полная копия в zip)", inc, []string{"файл", "размер"})
	os.WriteFile(filepath.Join(r.outDir, "report.html"), []byte(b.String()), 0o644)
	rp := filepath.Join(r.outDir, "report.html")
	r.p.set(func(p *Progress) { p.Report = rp })
	r.logf("Отчёт: %s", rp)
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
