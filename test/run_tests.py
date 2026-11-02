#!/usr/bin/env python3
"""End-to-end test: builds fixture, runs the Linux build, checks results with exiftool."""
import json, os, shutil, subprocess, sys, glob, time, urllib.request
H = os.path.dirname(os.path.abspath(__file__)); os.chdir(H)
BIN = sys.argv[1] if len(sys.argv) > 1 else "/tmp/tr"
subprocess.run([sys.executable, "make_fixture.py"], check=True, stdout=subprocess.DEVNULL)
SRC = ["-src", "fixture/D/Google Photo Export", "-src", "fixture/E/Takeout2023"]
fails = []
def check(c, msg):
    print(("PASS " if c else "FAIL ") + msg)
    if not c: fails.append(msg)
def run(dest, *extra):
    r = subprocess.run([BIN, *SRC, "-dest", dest, "-tz", "Europe/Minsk", *extra], capture_output=True, text=True)
    return r
def tags(p):
    return json.loads(subprocess.check_output(["exiftool", "-j", "-n", "-G1", "-api", "QuickTimeUTC=1", p]))[0]
shutil.rmtree("out", ignore_errors=True)
D = "out/Фото (восстановлено)"
r = run(D, "-dry-run"); check(r.returncode == 0 and not os.path.exists("out"), "dry-run writes nothing to destination")
r = run(D); check(r.returncode == 0, "real run exit 0"); print(r.stdout[-400:])
files = sorted(os.path.relpath(p, D) for p in glob.glob(D + "/**/*", recursive=True) if os.path.isfile(p) and "_restore" not in p)
exp = {"2019/2019-06/IMG_20190615_101010.jpg", "2021/2021-06/Screenshot_2021-06-29-13-43-47-185_by.st.priorm.jpg",
       "2021/2021-01/IMG_9999_some_long_descriptive_filename_here_x_full_original.jpg", "2020/2020-03/IMG_0001.jpg", "2020/2020-04/IMG_0001(1).jpg",
       "2020/2020-05/PXL_20200505_121212345.jpg", "2020/2020-05/PXL_20200505_121212345-edited.jpg", "2019/2019-08/IMG_1234.JPG", "2019/2019-08/IMG_1234.MP4",
       "2019/2019-01/IMG-20190101-WA0001.jpg", "2019/Unknown month/random.png", "2019/2019-07/VID_20190704_202020.mp4", "2018/2018-02/MVI_0001.AVI",
       "2018/2018-02/MVI_0001.AVI.xmp", "2018/2018-02/MVI_0001.THM", "2019/2019-09/anim.gif", "2017/2017-01/фото.jpg", "2017/2017-01/фото-изменено.jpg",
       "2015/2015-07/DSC_0001.JPG", "2014/2014-04/DSC_0002.JPG", "No date/blob.bmp", "No date/blob.bmp.xmp", "_Trash/deleted.jpg", "2017/2017-02/arch.jpg",
       "2022/2022-01/IMG_20220101_000001.jpg", "2022/2022-02/PXL_20220202_101010123.MP.jpg", "2022/2022-03/IMG_5555(1).jpg",
       "Albums/Archive/arch.jpg", "Albums/Германия-2017 (поездка)/IMG_20190615_101010.jpg", "Albums/Дети/DSC_0001.JPG", "Albums/Дети/DSC_0002.JPG",
       "Albums/Дети/blob.bmp", "Albums/Дети/фото-изменено.jpg", "Albums/Дети/фото.jpg", "Albums/Отпуск в Крыму/IMG_20220101_000001.jpg"}
check(set(files) == exp, f"folder structure exact ({len(files)} files)" + ("" if set(files) == exp else f" extra={set(files)-exp} missing={exp-set(files)}"))
st = lambda p: os.stat(os.path.join(D, p))
check(st("Albums/Дети/фото.jpg").st_ino == st("2017/2017-01/фото.jpg").st_ino, "album copy is a hard link (same inode)")
check(sum(1 for f in files if f.endswith("IMG_20190615_101010.jpg")) == 2, "dedupe: photo in 2 albums x 2 exports + extracted copy -> 1 file + 1 link")
t = tags(D + "/2019/2019-06/IMG_20190615_101010.jpg")
check(t.get("ExifIFD:DateTimeOriginal") == "2019:06:15 10:10:10" and t.get("ExifIFD:OffsetTimeOriginal") == "+03:00", "JPG DateTimeOriginal = UTC->local (+03:00)")
check(abs(t.get("Composite:GPSLatitude", 0) - 52.52) < 1e-4 and abs(t.get("Composite:GPSLongitude", 0) - 13.405) < 1e-4 and abs(t.get("Composite:GPSAltitude", 0) - 34.5) < 1e-3, "JPG GPS lat/lon/alt")
check(t.get("IFD0:ImageDescription") == "Отпуск на море" and t.get("IPTC:Caption-Abstract") == "Отпуск на море", "description (Cyrillic) in EXIF+IPTC+XMP")
check(t.get("XMP-xmp:Rating") == 5, "favorited -> Rating 5")
check("Вадим" in str(t.get("XMP-iptcExt:PersonInImage")) and "Германия-2017 (поездка)" in str(t.get("XMP-dc:Subject")), "people + album (from metadata.json title) keywords")
check(abs(os.stat(D + "/2019/2019-06/IMG_20190615_101010.jpg").st_mtime - 1560582610) < 2, "file mtime = photoTakenTime")
t = tags(D + "/2019/2019-07/VID_20190704_202020.mp4")
check(str(t.get("QuickTime:CreateDate", "")).startswith("2019:07:04") and abs(t.get("Composite:GPSLatitude", 0) - 41.9) < 1e-4, "MP4 QuickTime CreateDate + Keys GPS")
raw = json.loads(subprocess.check_output(["exiftool", "-j", "-QuickTime:CreateDate", D + "/2019/2019-07/VID_20190704_202020.mp4"]))[0]
check(raw.get("CreateDate") == "2019:07:04 17:20:20", "MP4 CreateDate stored in UTC")
t = tags(D + "/2019/2019-08/IMG_1234.MP4"); check(abs(t.get("Composite:GPSLatitude", 0) - 53.849) < 1e-4, "live-photo video got photo's json")
t = tags(D + "/2014/2014-04/DSC_0002.JPG"); check(t.get("ExifIFD:DateTimeOriginal") == "2014:04:04 04:04:04" and abs(t.get("Composite:GPSLatitude", 0) - 10) < 1e-4, "existing EXIF date kept, missing GPS filled")
t = tags(D + "/2015/2015-07/DSC_0001.JPG"); check(t.get("ExifIFD:DateTimeOriginal") == "2015:07:07 07:07:07", "no json -> date from existing EXIF")
t = tags(D + "/2019/2019-01/IMG-20190101-WA0001.jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2019:01:01 12:00:00", "no json -> date from filename (WhatsApp) written")
t = tags(D + "/2018/2018-02/MVI_0001.AVI.xmp"); check(str(t.get("XMP-exif:DateTimeOriginal", "")).startswith("2018:02:03 13:00:00"), "AVI -> XMP sidecar with date")
t = tags(D + "/2019/2019-09/anim.gif"); check(str(t.get("XMP-exif:DateTimeOriginal", "")).startswith("2019:09:09 12:09:09"), "GIF -> XMP date")
t = tags(D + "/2020/2020-05/PXL_20200505_121212345-edited.jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2020:05:05 12:12:12", "-edited takes original's json")
t = tags(D + "/2020/2020-04/IMG_0001(1).jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2020:04:02 15:00:00", "name(1).jpg <-> name.jpg(1).json")
t = tags(D + "/2022/2022-03/IMG_5555(1).jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2022:03:03 06:03:03", "supplemental-metadata(1).json")
t = tags(D + "/2022/2022-02/PXL_20220202_101010123.MP.jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2022:02:02 13:10:10", "truncated supplemental json")
t = tags(D + "/2021/2021-06/Screenshot_2021-06-29-13-43-47-185_by.st.priorm.jpg"); check(t.get("ExifIFD:DateTimeOriginal") == "2021:06:29 13:43:47", "truncated json name (real case)")
rep = open(D + "/_restore/report.html", encoding="utf-8").read()
check("Unmatched JSON (no media file) (1)" in rep and "ghost.jpg" in rep, "report lists unmatched json")
check("Skipped incomplete extracted copies (a full copy exists in a zip) (1)" in rep, "incomplete extracted copy detected and skipped")
check("status" in open(D + "/_restore/files.csv", encoding="utf-8-sig").readline(), "csv header is English")
check(os.path.exists(D + "/_restore/files.csv") and os.path.exists(D + "/_restore/state.json"), "files.csv + state.json")
stj = json.load(open(D + "/_restore/state.json", encoding="utf-8"))
check(stj.get("lang") == "en" and isinstance(stj.get("files"), dict) and len(stj["files"]) > 0, "state.json stores en folder language")
snap = lambda: sorted((os.path.relpath(p, D), os.stat(p).st_ino, os.stat(p).st_mtime, os.stat(p).st_size) for p in glob.glob(D + "/**/*", recursive=True) if os.path.isfile(p) and "_restore" not in p)
b = snap(); r = run(D); check(r.returncode == 0 and snap() == b and "copied: 0, already there: 25" in r.stdout, "re-run is idempotent")
os.remove(D + "/2017/2017-01/фото.jpg"); os.remove(D + "/Albums/Германия-2017 (поездка)/IMG_20190615_101010.jpg")
r = run(D); check("copied: 1, already there: 24" in r.stdout and os.path.exists(D + "/2017/2017-01/фото.jpg") and os.path.exists(D + "/Albums/Германия-2017 (поездка)/IMG_20190615_101010.jpg"), "resume: restores missing file and album link only")
# alternative options
D2 = "out/alt"; r = run(D2, "-layout", "ymd", "-albums", "copy", "-rename", "-skip-edited")
f2 = sorted(os.path.relpath(p, D2) for p in glob.glob(D2 + "/**/*", recursive=True) if os.path.isfile(p) and "_restore" not in p)
check("2019/2019-06-15/2019-06-15_101010_IMG_20190615_101010.jpg" in f2 and not any("edited" in f or "изменено" in f for f in f2), "layout ymd + rename + skip-edited")
check(os.stat(D2 + "/Albums/Дети/2017-01-01_003000_фото.jpg").st_ino != os.stat(D2 + "/2017/2017-01-01/2017-01-01_003000_фото.jpg").st_ino, "albums=copy makes real copies")
D3 = "out/none"; r = run(D3, "-albums", "none", "-layout", "y"); check(not os.path.exists(D3 + "/Albums") and os.path.exists(D3 + "/2019/IMG_1234.JPG"), "albums=none, layout y")
D4 = "out/album"; r = run(D4, "-layout", "album"); check(os.path.exists(D4 + "/Дети/фото.jpg") and os.path.exists(D4 + "/2019/IMG_1234.JPG"), "layout album")
for e in glob.glob("fixture/**/*", recursive=True): pass
check(subprocess.run(["unzip", "-tq", "fixture/D/Google Photo Export/takeout-20220224T200921Z-001.zip"], capture_output=True).returncode == 0, "sources untouched (zip intact)")
# Russian folder names, and a later English run must not split the tree.
Dru = "out/ru"
r = run(Dru, "-lang", "ru")
check(r.returncode == 0, "lang=ru exit 0")
check(os.path.exists(Dru + "/Альбомы/Дети/фото.jpg") and os.path.exists(Dru + "/Без даты/blob.bmp") and os.path.exists(Dru + "/_Корзина/deleted.jpg") and os.path.exists(Dru + "/2019/Месяц неизвестен/random.png") and os.path.exists(Dru + "/Альбомы/Архив/arch.jpg"), "lang=ru uses Russian folder names")
check("JSON без медиафайла (1)" in open(Dru + "/_restore/report.html", encoding="utf-8").read(), "lang=ru report")
check("скопировано:" in r.stdout, "lang=ru console summary")
stj = json.load(open(Dru + "/_restore/state.json", encoding="utf-8"))
check(stj.get("lang") == "ru", "state.json stores ru")
os.remove(Dru + "/2017/2017-01/фото.jpg")
r = run(Dru, "-lang", "en")
check(os.path.exists(Dru + "/2017/2017-01/фото.jpg") and os.path.exists(Dru + "/Альбомы") and not os.path.exists(Dru + "/Albums") and not os.path.exists(Dru + "/No date"), "re-run keeps stored Russian folder names")
check(json.load(open(Dru + "/_restore/state.json", encoding="utf-8")).get("lang") == "ru", "stored folder language unchanged")
check("Keeping them so the library is not split" in r.stdout, "log explains the locked folder language")
h = subprocess.run([BIN, "-h"], capture_output=True, text=True)
check("YYYY\\YYYY-MM" in h.stderr and "-lang" in h.stderr, "english -h")
hr = subprocess.run([BIN, "-lang", "ru", "-h"], capture_output=True, text=True)
check("ГГГГ\\ГГГГ-ММ" in hr.stderr, "russian -h")
miss = subprocess.run([BIN, "-src", SRC[1], "-dest", "out/no-exiftool", "-exiftool", "/no/such/exiftool", "-lang", "en"], capture_output=True, text=True)
check(miss.returncode != 0 and "sudo apt install libimage-exiftool-perl" in miss.stderr and "perl-Image-ExifTool" in miss.stderr and "perl-image-exiftool" in miss.stderr and "-exiftool" in miss.stderr, "missing exiftool install hint")
# Stop must interrupt a live run (HTTP API, same endpoint as the UI button)
# and leave a destination that a second run can finish.
def http_json(port, path, body=None, timeout=30):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request("http://127.0.0.1:%d%s" % (port, path), data=data, method="POST" if body is not None else "GET")
    if data is not None:
        req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode())
def descendants(pid):
    out = subprocess.run(["ps", "-o", "pid=,cmd=", "--ppid", str(pid)], capture_output=True, text=True).stdout
    kids = []
    for line in out.splitlines():
        line = line.strip()
        if not line:
            continue
        pid_s, _, cmd = line.partition(" ")
        try:
            cpid = int(pid_s)
        except ValueError:
            continue
        kids.append((cpid, cmd.strip()))
        kids.extend(descendants(cpid))
    return kids
stop_port = 18765
stop_body = {"sources": [SRC[1], SRC[3]], "dest": os.path.abspath("out/stop"), "layout": "ym", "albums": "hardlink",
             "tz": "Europe/Minsk", "lang": "en", "xmpSidecar": True, "rating": True, "albumKeywords": True, "people": True, "workers": 1}
srv = subprocess.Popen([BIN, "-port", str(stop_port), "-no-browser", "-lang", "en"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    up = False
    for _ in range(50):
        try:
            http_json(stop_port, "/api/status", timeout=0.5)
            up = True
            break
        except Exception:
            time.sleep(0.1)
    check(up, "stop-test server started")
    if up:
        http_json(stop_port, "/api/start", stop_body)
        seen = None
        for _ in range(400):
            seen = http_json(stop_port, "/api/status")
            if seen.get("running") and (seen.get("done") or 0) >= 1:
                break
            if seen.get("finished") and not seen.get("running"):
                break
            time.sleep(0.05)
        check(bool(seen and seen.get("running")), "run is still going when Stop is pressed")
        t0 = time.time()
        http_json(stop_port, "/api/cancel", {})
        st = seen
        while time.time() - t0 < 8:
            st = http_json(stop_port, "/api/status")
            if not st.get("running"):
                break
            time.sleep(0.1)
        elapsed = time.time() - t0
        check(not st.get("running") and elapsed < 5, "stop ends the run within 5s (%.2fs, phase=%s)" % (elapsed, st.get("phase")))
        check(st.get("stopped") is True and st.get("phase") == "Stopped" and not st.get("error"), "UI status is Stopped and Start can run again")
        time.sleep(0.3)
        ex_kids = [cmd for _, cmd in descendants(srv.pid) if "exiftool" in cmd.lower()]
        check(not ex_kids, "exiftool child processes were terminated (%s)" % ex_kids)
        copied = (st.get("stats") or {}).get("Copied") or 0
        if copied:
            sj = json.load(open("out/stop/_restore/state.json", encoding="utf-8"))
            check(sj.get("lang") == "en" and len(sj.get("files") or {}) >= 1, "state.json saved so a re-run can resume")
        http_json(stop_port, "/api/start", stop_body)
        fin = None
        t1 = time.time()
        while time.time() - t1 < 120:
            fin = http_json(stop_port, "/api/status")
            if fin.get("finished") and not fin.get("running"):
                break
            time.sleep(0.2)
        check(fin and fin.get("finished") and not fin.get("running") and not fin.get("error") and not fin.get("stopped") and fin.get("phase") == "Done", "re-run after stop completes")
        got = set(os.path.relpath(p, "out/stop") for p in glob.glob("out/stop/**/*", recursive=True) if os.path.isfile(p) and "_restore" not in p)
        check(got == exp, "re-run restores the full library" + ("" if got == exp else f" extra={got-exp} missing={exp-got}"))
finally:
    srv.terminate()
    try:
        srv.wait(timeout=5)
    except subprocess.TimeoutExpired:
        srv.kill()
print("\n%d failures" % len(fails)); sys.exit(1 if fails else 0)
