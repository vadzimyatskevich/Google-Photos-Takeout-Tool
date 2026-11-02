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
- **Simple web interface** in your browser, or a command line for scripting.

## Quick start

1. Download `TakeoutRestore.zip` from [Releases](../../releases) and unzip it.
2. Double-click `TakeoutRestore.exe`; a page opens in your browser.
3. Add your Takeout folder (for example `D:\Takeout`), choose an output folder, and run **Dry-run** first.
4. Check the report, then press **Start**.

> Windows SmartScreen may warn about an unsigned app: click **More info → Run anyway**.

## Command line

```
TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -dry-run
TakeoutRestore.exe -src "D:\Takeout" -src "E:\Takeout 2023" -dest "D:\Photos Restored"
```

Defaults: `-layout ym -albums hardlink -tz local -xmp -rating -album-keywords -people`. Use `-overwrite` to replace existing tags, `-tz Europe/Berlin` to set a time zone, and `-force` to start even if free space looks short. Run `TakeoutRestore.exe -h` for all options.

## Output

```
Photos Restored/
├── 2019/2019-07/IMG_20190714_181502.jpg
├── 2021/2021-05/IMG_20210524_110114.jpg
├── Альбомы/Germany 2017/…      (album folders, hard links)
├── Без даты/                   (no date found)
├── _Корзина/                   (Google Photos trash)
└── _restore/report.html, files.csv, log.txt, state.json
```

Folder names in the output are currently Russian; the UI is in Russian too. English localization is planned.

## How dates are chosen

1. `photoTakenTime` from the Takeout JSON (UTC, converted to the system time zone or `-tz`)
2. Existing EXIF `DateTimeOriginal`
3. Date in the file name
4. Year of the `Photos from YYYY` folder
5. Otherwise the file goes to the "no date" folder

Only missing tags are written unless `-overwrite` is set. File modified and creation times are set to the date taken.

## Building from source

Requires Go 1.21+, `curl` and `zip`. Pure Go, no external modules.

```
./build.sh        # downloads ExifTool for Windows, builds dist/TakeoutRestore.exe and dist/TakeoutRestore.zip
```

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
