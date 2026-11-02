TakeoutRestore — restore photos from Google Takeout
====================================================

TakeoutRestore reads Google Takeout archives (takeout-*.zip) as they are, without
extracting them first. It sorts photos and videos into YEAR\YEAR-MONTH folders and
writes the capture date, GPS coordinates, description, people, and albums from
Google's JSON files into the files themselves.
Your original archives and folders are NOT modified and NOT deleted.

The interface is in English by default. Choose a language from the dropdown at the
top of the window (the choice is remembered) or start with -lang en|ru|de|fr|es|pl|lt.
Folder names follow the language (Albums, No date, _Trash, Archive, Unknown month
in English). For a destination that was already created, those names stay as they
are — the choice is saved in _restore/state.json — so a later run does not split
the library across two naming schemes.

HOW TO USE (Windows)
1. Unzip this archive anywhere (for example, the desktop).
2. Double-click TakeoutRestore.exe. A desktop window opens (no browser required).
   Windows SmartScreen may warn that the app is unrecognized (it is not signed):
   click More info -> Run anyway.
   The first launch unpacks the built-in ExifTool (a few seconds).
   Current Windows 10/11 already include WebView2, which the window uses.
3. On step 1 (Sources), add the folder that contains your Takeout zips (or add
   zip files). The same photo from different exports, albums, and partly
   extracted folders is kept only once.
4. On step 2, choose a destination folder that is NOT inside the folder with the
   archives, and pick layout / album options.
5. On step 3, click Dry run first — nothing is written, but you can see how many
   files, duplicates, and how much disk space are needed. Open the report.
6. Then click Start restore. If you stop (or close the window), run it again with
   the same destination: it continues where it stopped and does not create duplicates.
   Optional: TakeoutRestore.exe -web opens the older localhost browser UI.

WHAT YOU GET
  D:\Photos Restored\2021\2021-05\IMG_20210524_110114.jpg   — photos by date
  D:\Photos Restored\Albums\Germany 2017\...                — your albums
  D:\Photos Restored\No date\                               — files with no date
  D:\Photos Restored\_Trash\                                — what was in the Google Photos trash
  D:\Photos Restored\_restore\report.html                   — the report (+ files.csv, log.txt)

With -lang ru the same folders are named Альбомы, Без даты, _Корзина, Архив,
and Месяц неизвестен.

DISK SPACE
Photos are copied once. Album folders are NTFS hard links: the same file is
visible in two places but takes space only once. That works when the destination
is on NTFS (C: and D: usually are) or, on Linux and macOS, on the same filesystem.
Deleting a file in an album folder leaves it in the year folder, and the other
way around. If you want real copies, choose "Separate copies".
You need about as much free space as the photos weigh. The program checks this
before it starts.

DATES AND TIME ZONE
Google stores the time in UTC; the program converts it to this computer's time
zone (you can set another, for example Europe/Minsk). A date or GPS that is
already in the file is not overwritten unless you turn that option on.
AVI and BMP get a .xmp file next to them. Every file's modified/created time
is set to the capture date.

LINUX
Install ExifTool first (it is not built into the Linux binary):
  Debian/Ubuntu: sudo apt install libimage-exiftool-perl
  Fedora:        sudo dnf install perl-Image-ExifTool
  Arch:          sudo pacman -S perl-image-exiftool
Unpack TakeoutRestore-linux-amd64.tar.gz, then:
  chmod +x TakeoutRestore
  ./TakeoutRestore
A desktop window opens (needs WebKitGTK). Fallback browser UI: ./TakeoutRestore -web
(folder buttons need zenity/kdialog, or paste the path).
  ./TakeoutRestore -src ~/Takeout -dest ~/Photos-Restored -dry-run
  ./TakeoutRestore -src ~/Takeout -dest ~/Photos-Restored -lang ru
Full release pack (GUI + CLI for all platforms): .\build.ps1   or   ./build.sh
Local Windows GUI+CLI: .\build.ps1 -Local
Manual desktop GUI: go build -tags desktop,production,gui -ldflags "-s -w -H windowsgui" -o TakeoutRestore-windows-amd64-gui.exe .
Manual CLI+web only:  go build -o TakeoutRestore .

COMMAND LINE (optional)
  TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -dry-run
  TakeoutRestore.exe -src "D:\Takeout" -dest "D:\Photos Restored" -lang ru
  All options: TakeoutRestore.exe -h
