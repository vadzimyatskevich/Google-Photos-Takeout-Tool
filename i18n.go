package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Lang is a UI language. Add a catalog entry in catalogs to support another one;
// missing keys fall back to English.
type Lang string

const (
	LangEN Lang = "en"
	LangRU Lang = "ru"
	LangDE Lang = "de"
	LangFR Lang = "fr"
	LangES Lang = "es"
	LangPL Lang = "pl"
	LangLT Lang = "lt"
)

// SupportedLangs is the UI / -lang order shown to users.
var SupportedLangs = []Lang{LangEN, LangRU, LangDE, LangFR, LangES, LangPL, LangLT}

// errNoPicker is returned when this OS has no folder dialog available.
var errNoPicker = errors.New("no folder dialog")

// Internal result codes. Reports and the CSV show the translated label.
const (
	statusCopied   = "copied"
	statusResumed  = "resumed"
	statusWillCopy = "will_copy"
	statusError    = "error"
)

// Folders are the destination directory names for one language.
type Folders struct {
	Albums       string
	NoDate       string
	Trash        string
	Archive      string
	UnknownMonth string
}

func (l Lang) Folders() Folders {
	if l == LangRU {
		return Folders{
			Albums:       "Альбомы",
			NoDate:       "Без даты",
			Trash:        "_Корзина",
			Archive:      "Архив",
			UnknownMonth: "Месяц неизвестен",
		}
	}
	return Folders{
		Albums:       "Albums",
		NoDate:       "No date",
		Trash:        "_Trash",
		Archive:      "Archive",
		UnknownMonth: "Unknown month",
	}
}

// NormalizeLang maps user input to a supported language. Anything else is English.
func NormalizeLang(s string) Lang {
	if l, ok := parseLang(s); ok {
		return l
	}
	return LangEN
}

func parseLang(s string) (Lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ru", "rus", "russian":
		return LangRU, true
	case "en", "eng", "english":
		return LangEN, true
	case "de", "deu", "ger", "german", "deutsch":
		return LangDE, true
	case "fr", "fra", "fre", "french", "français", "francais":
		return LangFR, true
	case "es", "spa", "spanish", "español", "espanol":
		return LangES, true
	case "pl", "pol", "polish", "polski":
		return LangPL, true
	case "lt", "lit", "lithuanian", "lietuvių", "lietuviu":
		return LangLT, true
	default:
		return "", false
	}
}

// peekLang reads -lang from args before flag.Parse so -h can be translated.
// The bool is true when the flag was present.
func peekLang(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if a == "-lang" || a == "--lang" {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "en", true
		}
		if strings.HasPrefix(a, "-lang=") || strings.HasPrefix(a, "--lang=") {
			return a[strings.IndexByte(a, '=')+1:], true
		}
	}
	return "en", false
}

// T returns the message for key. Format verbs use fmt.Sprintf when args are passed.
func (l Lang) T(key string, args ...any) string {
	s, ok := catalogs[l][key]
	if !ok {
		if s, ok = catalogs[LangEN][key]; !ok {
			return key
		}
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

func (l Lang) E(key string, args ...any) error {
	return errors.New(l.T(key, args...))
}

func (l Lang) statusText(code string) string {
	switch code {
	case statusCopied, statusResumed, statusWillCopy, statusError:
		return l.T("status." + code)
	default:
		return code
	}
}

func (l Lang) dateSrcText(src string) string {
	switch src {
	case "json", "exif", "filename", "year", "none":
		return l.T("datesrc." + src)
	default:
		return src
	}
}

// detectFolderLang reports which naming scheme is already present in dest.
func detectFolderLang(dest string) (Lang, bool) {
	enN := folderLangHits(dest, LangEN.Folders())
	ruN := folderLangHits(dest, LangRU.Folders())
	switch {
	case enN > ruN:
		return LangEN, true
	case ruN > enN:
		return LangRU, true
	default:
		return "", false
	}
}

func folderLangHits(dest string, f Folders) int {
	n := 0
	for _, name := range []string{f.Albums, f.NoDate, f.Trash} {
		if isDir(filepath.Join(dest, name)) {
			n++
		}
	}
	if isDir(filepath.Join(dest, f.Albums, f.Archive)) {
		n++
	}
	if hasSubdirNamed(dest, f.UnknownMonth) {
		n++
	}
	return n
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func hasSubdirNamed(dest, name string) bool {
	ents, err := os.ReadDir(dest)
	if err != nil {
		return false
	}
	for _, e := range ents {
		if e.IsDir() && isDir(filepath.Join(dest, e.Name(), name)) {
			return true
		}
	}
	return false
}

var catalogs = map[Lang]map[string]string{
	LangEN: enMsg,
	LangRU: ruMsg,
	LangDE: deMsg,
	LangFR: frMsg,
	LangES: esMsg,
	LangPL: plMsg,
	LangLT: ltMsg,
}

var enMsg = map[string]string{
	"usage.header": "TakeoutRestore restores Google Photos Takeout metadata and organizes photos.\n\nUsage: TakeoutRestore [options] [sources...]",

	"flag.src":            "source: a folder or zip (repeat for more than one)",
	"flag.dest":           "destination folder (console mode when set)",
	"flag.layout":         "layout: ym (YYYY\\YYYY-MM) | y_m (YYYY\\MM) | ymd (YYYY\\YYYY-MM-DD) | y (YYYY) | album",
	"flag.albums":         "album folders: hardlink | copy | none",
	"flag.rename":         "rename files to YYYY-MM-DD_HHMMSS_name",
	"flag.tz":             "time zone for dates: local | UTC | Europe/Minsk | +03:00",
	"flag.overwrite":      "overwrite existing tags (default: fill empty tags only)",
	"flag.skip-edited":    "skip edited copies (-edited)",
	"flag.xmp":            "write an XMP sidecar for avi/bmp and when the file itself cannot be updated",
	"flag.rating":         "favorites -> rating 5",
	"flag.album-keywords": "album names -> XMP keywords",
	"flag.people":         "people names -> keywords / PersonInImage",
	"flag.dry-run":        "dry run, write nothing to the destination",
	"flag.force":          "continue even when free disk space looks short",
	"flag.workers":        "number of threads (0 = auto)",
	"flag.exiftool":       "path to exiftool (Windows: built-in; Linux/macOS: on PATH)",
	"flag.port":           "web interface port (0 = random)",
	"flag.no-browser":     "do not open a browser",
	"flag.web":            "use the localhost browser UI instead of the desktop window",
	"flag.lang":           "language: en | ru | de | fr | es | pl | lt (default en)",

	"cli.summary": "\nUnique: %d, duplicates: %d, copied: %d, already there: %d, errors: %d\nDates: json %d, exif %d, filename %d, year %d, none %d; GPS %d\n",
	"cli.report":  "Report: %s",
	"cli.error":   "ERROR: %v",

	"web.port":     "Could not open a port: %v",
	"web.banner":   "TakeoutRestore — restore Google Photos from a Takeout export",
	"web.url":      "Open this address in a browser: %s",
	"web.openfail": "Could not open a browser automatically. Use the address above.",
	"web.keepopen": "Leave this window open while TakeoutRestore is running. Close the window to quit.",
	"web.running":  "A run is already in progress",
	"web.nofile":   "file not found",

	"phase.prep":     "Preparing",
	"phase.error":    "Error",
	"phase.done":     "Done",
	"phase.stopping": "Stopping…",
	"phase.stopped":  "Stopped",
	"phase.scan":     "Scanning sources",
	"phase.check":    "Checking extracted files",
	"phase.json":     "Reading JSON metadata",
	"phase.copy":     "Copying and writing metadata",
	"phase.dry":      "Dry run (nothing is written)",
	"phase.paused":   "Paused",

	"progress.copy":      "copying…",
	"progress.meta.read": "reading metadata…",
	"progress.meta.write": "writing metadata…",

	"err.cancelled":  "stopped by the user",
	"err.internal":   "internal error: %v",
	"err.nodest":     "destination folder is not set",
	"err.nosrc":      "no sources given",
	"err.tz":         "time zone %q: %v",
	"err.destinside": "the destination cannot be inside a source: %s",
	"err.exiftool":   "ExifTool was not found (%v).\nInstall it, or pass -exiftool with the path to the executable.\n  Debian/Ubuntu: sudo apt install libimage-exiftool-perl\n  Fedora: sudo dnf install perl-Image-ExifTool\n  Arch: sudo pacman -S perl-image-exiftool\n  macOS: brew install exiftool",
	"err.exifembed":  "could not unpack the built-in ExifTool: %v. Pass -exiftool with a path to exiftool.",
	"err.nomedia":    "no photos or videos found in the sources",
	"err.diskshort":  "Not enough disk space: need about %.1f GB, free %.1f GB",
	"err.diskmore":   "%s. Free some space, or turn on “continue anyway”.",
	"err.exifstart":  "could not start exiftool (%s): %v",
	"err.size":       "size %d instead of %d",
	"err.read":       "read: %v",
	"err.mkdir":      "create folder: %v",
	"err.move":       "move: %v",
	"err.album":      "album %s: %v",
	"err.exifline":   "exiftool: %s",

	"log.start":      "=== Start (dry-run=%v) sources=%q dest=%q layout=%s albums=%s tz=%s lang=%s",
	"log.state":      "Previous run state: %d files",
	"log.exiftool":   "exiftool: %s",
	"log.warn":       "WARNING: %s",
	"log.found":      "Found media: %d, json: %d, zip: %d",
	"log.readerr":    "ERROR reading %s: %v",
	"log.unique":     "Unique files: %d (%.2f GB), duplicates: %d, incomplete copies skipped: %d",
	"log.done":       "=== Finished. Copied %d, already there %d, errors %d",
	"log.error":      "ERROR: %v",
	"log.errfile":    "ERROR [%s]: %s",
	"log.mtime":      "could not set the file time for %s: %v",
	"log.hardlink":   "hard link failed (%v), copying: %s",
	"log.report":     "Report: %s",
	"log.stopped":    "Stopped by the user. Files already written are kept; run again to continue.",
	"log.lock.ru": "This destination already uses Russian folder names (Альбомы, Без даты, _Корзина, Архив, Месяц неизвестен). Keeping them so the library is not split. Messages and the report use the selected language.",
	"log.lock.en": "This destination already uses English folder names (Albums, No date, _Trash, Archive, Unknown month). Keeping them so the library is not split. Messages and the report use the selected language.",

	"status.copied":    "copied",
	"status.resumed":   "already there",
	"status.will_copy": "will be copied",
	"status.error":     "error",

	"datesrc.json":     "JSON",
	"datesrc.exif":     "EXIF",
	"datesrc.filename": "file name",
	"datesrc.year":     "year only",
	"datesrc.none":     "none",

	"pick.title":     "Select a folder",
	"pick.err":       "A folder dialog is not available. Paste the path instead.",
	"pick.errdetail": "Could not open a folder dialog: %v. Paste the path instead.",

	"rep.title":   "Restore report",
	"rep.dry":     " — DRY RUN (nothing was written)",
	"rep.timefmt": "2006-01-02 15:04",
	"rep.meta":    "Destination: <b>%s</b><br>Sources: %s<br>Time zone: %s",
	"rep.footer":  "Full file list: <code>files.csv</code> (opens in Excel), log: <code>log.txt</code>.",
	"rep.gb":      "%.2f GB",
	"rep.gbpair":  "%.1f / %.1f GB",

	"rep.media":      "Media files found in the sources (every copy)",
	"rep.unique":     "Unique files",
	"rep.ubytes":     "Size of unique files",
	"rep.dups":       "Duplicates skipped (same file in another album, archive, or export)",
	"rep.incomplete": "Incomplete or broken copies in folders skipped",
	"rep.edited":     "Edited versions skipped",
	"rep.copied":     "Copied",
	"rep.resumed":    "Already there (run resumed)",
	"rep.datejson":   "Date from JSON (Google)",
	"rep.dateexif":   "Date from EXIF",
	"rep.datename":   "Date from the file name",
	"rep.dateyear":   "Year only (“Photos from YYYY” folder)",
	"rep.datenone":   "No date",
	"rep.metafiles":  "Files with metadata written",
	"rep.gps":        "GPS written",
	"rep.desc":       "Description written",
	"rep.xmp":        "XMP sidecars (avi, bmp, …)",
	"rep.links":      "Album hard links",
	"rep.copies":     "Album copies",
	"rep.trash":      "In “%s”",
	"rep.jsontotal":  "JSON metadata files",
	"rep.unmatched":  "JSON without a media file",
	"rep.nojson":     "Files without JSON",
	"rep.errors":     "Errors",
	"rep.space":      "Disk space needed / free",

	"rep.sec.errors":     "Errors",
	"rep.sec.nojson":     "Files without JSON",
	"rep.sec.unmatched":  "Unmatched JSON (no media file)",
	"rep.sec.incomplete": "Skipped incomplete extracted copies (a full copy exists in a zip)",
	"rep.h.file":         "file",
	"rep.h.error":        "error",
	"rep.h.datesrc":      "date source",
	"rep.h.dest":         "destination",
	"rep.h.json":         "json",
	"rep.h.title":        "title",
	"rep.h.size":         "size",

	"csv.status":  "status",
	"csv.file":    "file",
	"csv.date":    "date",
	"csv.datesrc": "date source",
	"csv.json":    "json",
	"csv.match":   "json match",
	"csv.albums":  "albums",
	"csv.copies":  "copies in sources",
	"csv.source":  "source",
	"csv.error":   "error",
}

var ruMsg = map[string]string{
	"usage.header": "TakeoutRestore восстанавливает метаданные Google Фото из Takeout и раскладывает файлы по папкам.\n\nИспользование: TakeoutRestore [параметры] [источники...]",

	"flag.src":            "источник: папка или zip (можно несколько раз)",
	"flag.dest":           "папка назначения (если указана — консольный режим)",
	"flag.layout":         "структура: ym (ГГГГ\\ГГГГ-ММ) | y_m (ГГГГ\\ММ) | ymd (ГГГГ\\ГГГГ-ММ-ДД) | y (ГГГГ) | album",
	"flag.albums":         "папки альбомов: hardlink | copy | none",
	"flag.rename":         "переименовать в ГГГГ-ММ-ДД_ЧЧММСС_имя",
	"flag.tz":             "часовой пояс для дат: local | UTC | Europe/Minsk | +03:00",
	"flag.overwrite":      "перезаписывать существующие теги (по умолчанию только пустые)",
	"flag.skip-edited":    "пропускать отредактированные копии (-edited/-изменено)",
	"flag.xmp":            "XMP-файл рядом для avi/bmp и при ошибке записи",
	"flag.rating":         "избранное -> рейтинг 5",
	"flag.album-keywords": "названия альбомов -> ключевые слова XMP",
	"flag.people":         "имена людей -> ключевые слова/PersonInImage",
	"flag.dry-run":        "пробный прогон, ничего не записывать",
	"flag.force":          "продолжать даже при нехватке места",
	"flag.workers":        "число потоков (0 = авто)",
	"flag.exiftool":       "путь к exiftool (Windows: встроенный; Linux/macOS: из PATH)",
	"flag.port":           "порт веб-интерфейса (0 = случайный)",
	"flag.no-browser":     "не открывать браузер",
	"flag.web":            "открыть веб-интерфейс в браузере вместо окна приложения",
	"flag.lang":           "язык: en | ru | de | fr | es | pl | lt (по умолчанию en)",

	"cli.summary": "\nУникальных: %d, дубликатов: %d, скопировано: %d, уже было: %d, ошибок: %d\nДаты: json %d, exif %d, имя файла %d, год %d, нет %d; GPS %d\n",
	"cli.report":  "Отчёт: %s",
	"cli.error":   "ОШИБКА: %v",

	"web.port":     "Не удалось открыть порт: %v",
	"web.banner":   "TakeoutRestore — восстановление Google Фото из Takeout",
	"web.url":      "Откройте этот адрес в браузере: %s",
	"web.openfail": "Не удалось открыть браузер автоматически. Откройте адрес выше вручную.",
	"web.keepopen": "Не закрывайте это окно, пока идёт работа. Закрыть программу — закрыть это окно.",
	"web.running":  "уже выполняется",
	"web.nofile":   "нет файла",

	"phase.prep":     "Подготовка",
	"phase.error":    "Ошибка",
	"phase.done":     "Готово",
	"phase.stopping": "Остановка…",
	"phase.stopped":  "Остановлено",
	"phase.scan":     "Сканирование источников",
	"phase.check":    "Проверка распакованных файлов",
	"phase.json":     "Чтение JSON-метаданных",
	"phase.copy":     "Копирование и запись метаданных",
	"phase.dry":      "Пробный прогон (ничего не записывается)",
	"phase.paused":   "Пауза",

	"progress.copy":       "копирование…",
	"progress.meta.read":  "чтение метаданных…",
	"progress.meta.write": "запись метаданных…",

	"err.cancelled":  "остановлено пользователем",
	"err.internal":   "внутренняя ошибка: %v",
	"err.nodest":     "не указана папка назначения",
	"err.nosrc":      "не указаны источники",
	"err.tz":         "часовой пояс %q: %v",
	"err.destinside": "папка назначения не может быть внутри источника: %s",
	"err.exiftool":   "ExifTool не найден (%v).\nУстановите его или укажите путь флагом -exiftool.\n  Debian/Ubuntu: sudo apt install libimage-exiftool-perl\n  Fedora: sudo dnf install perl-Image-ExifTool\n  Arch: sudo pacman -S perl-image-exiftool\n  macOS: brew install exiftool",
	"err.exifembed":  "не удалось распаковать встроенный ExifTool: %v. Укажите путь флагом -exiftool.",
	"err.nomedia":    "в источниках не найдено фото/видео",
	"err.diskshort":  "Недостаточно места: нужно ≈%.1f ГБ, свободно %.1f ГБ",
	"err.diskmore":   "%s. Освободите место или включите «всё равно продолжить».",
	"err.exifstart":  "не удалось запустить exiftool (%s): %v",
	"err.size":       "размер %d вместо %d",
	"err.read":       "чтение: %v",
	"err.mkdir":      "создание папки: %v",
	"err.move":       "перемещение: %v",
	"err.album":      "альбом %s: %v",
	"err.exifline":   "exiftool: %s",

	"log.start":      "=== Запуск (dry-run=%v) источники=%q назначение=%q структура=%s альбомы=%s TZ=%s lang=%s",
	"log.state":      "Найдено состояние прошлого запуска: %d файлов",
	"log.exiftool":   "exiftool: %s",
	"log.warn":       "ПРЕДУПРЕЖДЕНИЕ: %s",
	"log.found":      "Найдено медиа: %d, json: %d, zip: %d",
	"log.readerr":    "ОШИБКА чтения %s: %v",
	"log.unique":     "Уникальных файлов: %d (%.2f ГБ), дубликатов: %d, недораспакованных копий пропущено: %d",
	"log.done":       "=== Завершено. Скопировано %d, уже было %d, ошибок %d",
	"log.error":      "ОШИБКА: %v",
	"log.errfile":    "ОШИБКА [%s]: %s",
	"log.mtime":      "не удалось установить время файла %s: %v",
	"log.hardlink":   "жёсткая ссылка не удалась (%v), копирую: %s",
	"log.report":     "Отчёт: %s",
	"log.stopped":    "Остановлено пользователем. Уже записанные файлы сохранены; повторный запуск продолжит.",
	"log.lock.ru": "Эта папка назначения уже использует русские названия (Альбомы, Без даты, _Корзина, Архив, Месяц неизвестен). Оставляем их, чтобы библиотека не разделилась. Сообщения и отчёт — на выбранном языке.",
	"log.lock.en": "Эта папка назначения уже использует английские названия (Albums, No date, _Trash, Archive, Unknown month). Оставляем их, чтобы библиотека не разделилась. Сообщения и отчёт — на выбранном языке.",

	"status.copied":    "скопирован",
	"status.resumed":   "уже было",
	"status.will_copy": "будет скопирован",
	"status.error":     "ошибка",

	"datesrc.json":     "JSON",
	"datesrc.exif":     "EXIF",
	"datesrc.filename": "имя файла",
	"datesrc.year":     "только год",
	"datesrc.none":     "нет",

	"pick.title":     "Выберите папку",
	"pick.err":       "Окно выбора папки недоступно. Вставьте путь вручную.",
	"pick.errdetail": "Не удалось открыть окно выбора папки: %v. Вставьте путь вручную.",

	"rep.title":   "Отчёт о восстановлении",
	"rep.dry":     " — ПРОБНЫЙ ПРОГОН (ничего не записано)",
	"rep.timefmt": "02.01.2006 15:04",
	"rep.meta":    "Назначение: <b>%s</b><br>Источники: %s<br>Часовой пояс: %s",
	"rep.footer":  "Полный список файлов: <code>files.csv</code> (открывается в Excel), журнал: <code>log.txt</code>.",
	"rep.gb":      "%.2f ГБ",
	"rep.gbpair":  "%.1f / %.1f ГБ",

	"rep.media":      "Медиафайлов найдено в источниках (всех копий)",
	"rep.unique":     "Уникальных файлов",
	"rep.ubytes":     "Объём уникальных файлов",
	"rep.dups":       "Дубликатов пропущено (тот же файл в другом альбоме/архиве/выгрузке)",
	"rep.incomplete": "Недораспакованных/битых копий в папках пропущено",
	"rep.edited":     "Отредактированных версий пропущено",
	"rep.copied":     "Скопировано",
	"rep.resumed":    "Уже было (повторный запуск)",
	"rep.datejson":   "Дата из JSON (Google)",
	"rep.dateexif":   "Дата из EXIF файла",
	"rep.datename":   "Дата из имени файла",
	"rep.dateyear":   "Только год (папка «Photos from ГГГГ»)",
	"rep.datenone":   "Без даты",
	"rep.metafiles":  "Записаны метаданные (файлов)",
	"rep.gps":        "Записан GPS",
	"rep.desc":       "Записано описание",
	"rep.xmp":        "XMP-файлов рядом (avi, bmp…)",
	"rep.links":      "Ссылок в альбомах (жёсткие ссылки)",
	"rep.copies":     "Копий в альбомах",
	"rep.trash":      "В «%s»",
	"rep.jsontotal":  "JSON-метаданных",
	"rep.unmatched":  "JSON без файла",
	"rep.nojson":     "Файлов без JSON",
	"rep.errors":     "Ошибок",
	"rep.space":      "Нужно места / свободно",

	"rep.sec.errors":     "Ошибки",
	"rep.sec.nojson":     "Файлы без JSON",
	"rep.sec.unmatched":  "JSON без медиафайла",
	"rep.sec.incomplete": "Пропущенные недораспакованные копии (есть полная копия в zip)",
	"rep.h.file":         "файл",
	"rep.h.error":        "ошибка",
	"rep.h.datesrc":      "источник даты",
	"rep.h.dest":         "куда",
	"rep.h.json":         "json",
	"rep.h.title":        "title",
	"rep.h.size":         "размер",

	"csv.status":  "статус",
	"csv.file":    "файл",
	"csv.date":    "дата",
	"csv.datesrc": "источник даты",
	"csv.json":    "json",
	"csv.match":   "сопоставление json",
	"csv.albums":  "альбомы",
	"csv.copies":  "копий в источниках",
	"csv.source":  "источник",
	"csv.error":   "ошибка",
}
