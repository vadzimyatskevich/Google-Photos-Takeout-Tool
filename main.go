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
	o := defaultOptions()
	var srcs multiFlag
	flag.Var(&srcs, "src", "источник: папка или zip (можно несколько раз)")
	flag.StringVar(&o.Dest, "dest", "", "папка назначения (если указана — консольный режим)")
	flag.StringVar(&o.Layout, "layout", o.Layout, "структура: ym (ГГГГ\\ГГГГ-ММ) | y_m (ГГГГ\\ММ) | ymd (ГГГГ\\ГГГГ-ММ-ДД) | y (ГГГГ) | album")
	flag.StringVar(&o.Albums, "albums", o.Albums, "папки альбомов: hardlink | copy | none")
	flag.BoolVar(&o.Rename, "rename", false, "переименовать в ГГГГ-ММ-ДД_ЧЧММСС_имя")
	flag.StringVar(&o.TZ, "tz", o.TZ, "часовой пояс для дат: local | UTC | Europe/Minsk | +03:00")
	flag.BoolVar(&o.Overwrite, "overwrite", false, "перезаписывать существующие теги (по умолчанию только пустые)")
	flag.BoolVar(&o.SkipEdited, "skip-edited", false, "пропускать отредактированные копии (-edited/-изменено)")
	flag.BoolVar(&o.XMPSidecar, "xmp", o.XMPSidecar, "XMP-файл рядом для avi/bmp и при ошибке записи")
	flag.BoolVar(&o.Rating, "rating", o.Rating, "избранное -> рейтинг 5")
	flag.BoolVar(&o.AlbumKeywords, "album-keywords", o.AlbumKeywords, "названия альбомов -> ключевые слова XMP")
	flag.BoolVar(&o.People, "people", o.People, "имена людей -> ключевые слова/PersonInImage")
	flag.BoolVar(&o.DryRun, "dry-run", false, "пробный прогон, ничего не записывать")
	flag.BoolVar(&o.Force, "force", false, "продолжать даже при нехватке места")
	flag.IntVar(&o.Workers, "workers", 0, "число потоков (0 = авто)")
	flag.StringVar(&o.ExifTool, "exiftool", "", "путь к exiftool (по умолчанию встроенный)")
	port := flag.Int("port", 0, "порт веб-интерфейса (0 = случайный)")
	noBrowser := flag.Bool("no-browser", false, "не открывать браузер")
	flag.Parse()
	o.Sources = append(srcs, flag.Args()...)
	if o.Dest != "" {
		cliMode = true
		p := &Progress{}
		err := Run(o, p)
		snap := p.snapshot()
		if s := p.Stats; s != nil {
			fmt.Printf("\nУникальных: %d, дубликатов: %d, скопировано: %d, уже было: %d, ошибок: %d\nДаты: json %d, exif %d, имя файла %d, год %d, нет %d; GPS %d\n",
				s.Unique, s.Duplicates, s.Copied, s.Resumed, s.Errors, s.DateJSON, s.DateEXIF, s.DateName, s.DateYear, s.DateNone, s.GPSWritten)
		}
		fmt.Println("Отчёт:", snap["report"])
		if err != nil {
			fmt.Fprintln(os.Stderr, "ОШИБКА:", err)
			os.Exit(1)
		}
		return
	}
	serveUI(*port, !*noBrowser)
}
