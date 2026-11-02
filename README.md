# TakeoutRestore: Google Photos Takeout Metadata Fixer & Photo Organizer

**Fix Google Takeout photos with wrong dates and missing GPS.** TakeoutRestore reads your Google Photos Takeout export (zip archives or extracted folders), writes the original date taken, GPS location and descriptions from Google's `.json` sidecar files back into your photos and videos, removes duplicates, and sorts everything into a clean `Year/Month` folder structure with separate album folders.

One portable Windows `.exe`. No Python, no installation, ExifTool built in.

> **Status: beta.** Tested end to end on a synthetic Takeout fixture. Run a **Dry-run** first and check the report before processing your real library.

## Why

When you export your library with Google Takeout, the photos come out with:
- the wrong "date taken" (often the export date),
- GPS and descriptions stored in separate `.json` files instead of in the image,
- the same photo duplicated in `Photos from YYYY` and in every album,
- JSON file names truncated or mismatched (`name.jpg(1).json`, `.supplemental-metadata.json`, `-edited`),
- a 50 GB export split across dozens of 2 GB zip files.

TakeoutRestore fixes all of that in one pass.

## Features

- **Restores metadata from Google Takeout JSON:** DateTimeOriginal, CreateDate, time zone offset, GPS latitude/longitude/altitude, description/caption, favorites (rating) and people names, written into EXIF, XMP and IPTC.
- **Videos too:** MP4, MOV and 3GP get QuickTime creation dates and GPS. AVI, BMP and other formats get an `.xmp` sidecar.
- **Reads zip files directly:** point it at the folder with `takeout-*.zip` files; nothing has to be extracted first. Handles albums split across multiple zip parts, Cyrillic and other non-Latin file names, and detects broken partial extractions.
- **Smart JSON matching:** truncated names, `(1)` duplicates, `-edited` copies, `.supplemental-metadata.json` (newer export format), Live Photos / motion photos, and RAW companions (CR2 + JPG, AVI + THM).
- **Fallbacks when JSON is missing:** existing EXIF, then the date in the file name (`IMG_20210524_110114`, `PXL_…`, `Screenshot_…`, WhatsApp), then the `Photos from YYYY` folder.
- **Duplicate removal** across albums and across multiple Takeout exports, keeping one copy.
- **Clean folder structure:** `2021/2021-05/…`, plus album folders made with NTFS hard links, so albums take no extra disk space.
- **Safe:** never modifies your source files. Dry-run mode previews everything, re-runs continue where they stopped, and an HTML report lists every file, duplicate and unmatched JSON.
- **Simple web interface** in your browser (English by default, Russian optional), or a command line for scripting.

## Quick start

1. Download `TakeoutRestore.zip` from [Releases](../../releases) and unzip it.
2. Double-click `TakeoutRestore.exe`; a page opens in your browser. The interface is English. Switch to Russian with **RU** at the top of the page (the choice is remembered) or start with `-lang ru`.
3. Add your Takeout folder (for example `D:\Takeout`), choose an output folder, and run **Dry-run** first.
4. Check the report, then press **Start**.

> Windows SmartScreen may warn about an unsigned app: click **More info → Run anyway**.

## Command line

```
TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -dry-run
TakeoutRestore.exe -src "D:\Takeout" -src "E:\Takeout 2023" -dest "D:\Photos Restored"
TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -lang ru
```

Defaults: `-layout ym -albums hardlink -tz local -xmp -rating -album-keywords -people -lang en`. Use `-overwrite` to replace existing tags, `-tz Europe/Berlin` to set a time zone, and `-force` to start even if free space looks short. `-lang en|ru` selects the language of the interface, the report, and (for a new destination) the folder names. Run `TakeoutRestore.exe -h` for all options. Add `-lang ru` before `-h` to read the help in Russian.

## Output

```
Photos Restored/
├── 2019/2019-07/IMG_20190714_181502.jpg
├── 2021/2021-05/IMG_20210524_110114.jpg
├── Albums/Germany 2017/…       (album folders, hard links)
├── No date/                    (no date found)
├── _Trash/                     (Google Photos trash)
└── _restore/report.html, files.csv, log.txt, state.json
```

Those folder names are English by default. With `-lang ru` (or **RU** in the web UI) the same folders are named `Альбомы`, `Без даты`, `_Корзина`, `Архив`, and `Месяц неизвестен`. The folder language is stored in `_restore/state.json`. A later run on that destination keeps the saved names, even if you switch language, so the library is not split across two schemes. Messages and the report follow the language of the current run.

## How dates are chosen

1. `photoTakenTime` from the Takeout JSON (UTC, converted to the system time zone or `-tz`)
2. Existing EXIF `DateTimeOriginal`
3. Date in the file name
4. Year of the `Photos from YYYY` folder
5. Otherwise the file goes to the "no date" folder

Only missing tags are written unless `-overwrite` is set. File modified and creation times are set to the date taken.

## Linux

The Linux build uses ExifTool from the system (it is not embedded).

1. Install ExifTool:
   - Debian/Ubuntu: `sudo apt install libimage-exiftool-perl`
   - Fedora: `sudo dnf install perl-Image-ExifTool`
   - Arch: `sudo pacman -S perl-image-exiftool`
2. Download `TakeoutRestore-linux-amd64.tar.gz` (or `TakeoutRestore-linux-arm64.tar.gz`), unpack it, and `chmod +x TakeoutRestore`.
3. Web UI: `./TakeoutRestore`. It opens a browser with `xdg-open`, or prints the URL if that fails. There is no Windows folder dialog: paste paths, or install `zenity` or `kdialog` to use the folder buttons.
4. Command line:

```
./TakeoutRestore -src ~/Takeout -dest ~/Photos-Restored -dry-run
./TakeoutRestore -src ~/Takeout -dest ~/Photos-Restored -lang ru
```

Album hard links work when the destination is on the same filesystem. If a link cannot be created, the file is copied instead.

Build just the Linux binary from source (Go 1.21+, no external modules):

```
go build -o TakeoutRestore .
```

macOS archives (`TakeoutRestore-darwin-amd64.tar.gz`, `TakeoutRestore-darwin-arm64.tar.gz`) work the same way. Install ExifTool with `brew install exiftool`. The browser opens with `open`.

## Building from source

Requires Go 1.21+, `curl` and `zip`. Pure Go, no external modules.

```
./build.sh
```

`./build.sh` downloads ExifTool for Windows and writes:

- `dist/TakeoutRestore.exe` and `dist/TakeoutRestore.zip` (Windows x64, ExifTool embedded; the zip includes `README.txt` and `ПРОЧТИ.txt`)
- `dist/TakeoutRestore-linux-amd64.tar.gz` and `dist/TakeoutRestore-linux-arm64.tar.gz`
- `dist/TakeoutRestore-darwin-amd64.tar.gz` and `dist/TakeoutRestore-darwin-arm64.tar.gz`

Tests (Linux, needs Python 3, ffmpeg, Pillow and exiftool):

```
go build -o /tmp/tr . && python3 test/run_tests.py /tmp/tr
```

## Known limitations

- Duplicates are detected by file size plus the CRC32 stored in the zip, not a cryptographic hash.
- Line breaks in descriptions are saved as spaces.
- An edited copy finds its original's JSON only when both have the same extension.
- Album hard links share one physical file: editing a photo in an album folder changes it in the year folder too.

## License

[MIT](LICENSE). The Windows build embeds [ExifTool](https://exiftool.org/) by Phil Harvey, distributed under its own license; see [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md).

## Keywords

Google Takeout photos wrong date, fix Google Photos metadata, Google Takeout JSON to EXIF, restore photo date taken, Google Photos export organizer, merge Takeout zip files, Google Photos Takeout helper alternative.
