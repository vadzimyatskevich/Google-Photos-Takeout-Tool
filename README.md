# TakeoutRestore — Fix Google Takeout Photos

**TakeoutRestore** is a free desktop tool that restores **Google Photos Takeout** exports: it fixes wrong dates, writes GPS and descriptions from Google’s JSON sidecars into EXIF, removes album duplicates, and sorts everything into a clean `Year/Month` library — reading `takeout-*.zip` files directly, without extracting them first.

One portable app for Windows, macOS and Linux. No Python. ExifTool is built into the Windows build.

![TakeoutRestore — restore Google Takeout photos with live preview](docs/screenshot.png)

> **Beta.** Run a **Dry run** first and check the report before processing your real library.

## Quick start

1. Download **TakeoutRestore** from [Releases](../../releases) and unzip it.
2. Open `TakeoutRestore.exe` (Windows) or `./TakeoutRestore` (macOS / Linux).
3. Add the folder with your Google Takeout zips, choose a destination, run **Dry run**.
4. Check the report, then press **Start restore**.

Windows SmartScreen may warn about an unsigned app: **More info → Run anyway**. WebView2 is required (included with current Windows 10/11). On Linux, install ExifTool and WebKitGTK first — see the [full guide](docs/USAGE.md#linux).

```
TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -dry-run
```

## What it fixes in Google Takeout

| Problem after Google Takeout | What TakeoutRestore does |
|---|---|
| Wrong “date taken” (often the export date) | Restores `photoTakenTime` into EXIF / QuickTime |
| GPS & captions stuck in `.json` files | Writes location, description, people, favorites into the media |
| Same photo in albums and `Photos from YYYY` | Keeps one copy; albums via hard links (no extra disk use) |
| Dozens of split `takeout-*.zip` parts | Reads zips in place; matches truncated / `(1)` / `-edited` JSON names |

## Languages

English by default. Switch language in the app (remembered) or with `-lang en|ru|…`. Folder names follow the language for a new destination.

## Docs & license

- **[Full usage guide](docs/USAGE.md)** — CLI options, output layout, date rules, Linux/macOS, building from source, limitations
- [MIT](LICENSE) · Embedded ExifTool: [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md)

---

*Keywords: Google Takeout photos, fix Google Photos Takeout metadata, restore date taken from Takeout JSON, Google Photos export organizer, merge Takeout zip files, Google Takeout EXIF GPS*
