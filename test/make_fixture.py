#!/usr/bin/env python3
"""Builds a synthetic Google Takeout fixture in ./fixture (two exports, split zips, quirks)."""
import json, os, shutil, subprocess, zipfile, sys
from datetime import datetime, timezone

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fixture")
shutil.rmtree(ROOT, ignore_errors=True)
MEDIA = os.path.join(ROOT, "_media"); os.makedirs(MEDIA)
n = [0]

def ff(*args):
    subprocess.run(["ffmpeg", "-loglevel", "error", "-y", *args], check=True)

def img(ext, seed):
    n[0] += 1
    p = os.path.join(MEDIA, f"m{n[0]}.{ext}")
    ff("-f", "lavfi", "-i", f"testsrc=size=160x120:rate=1:duration=1:decimals=0,drawtext=text='{seed}':fontsize=30:x=10:y=10:fontcolor=white" if False else f"color=c=0x{(seed*2654435761)&0xffffff:06x}:size=160x120:d=1", "-frames:v", "1", *(["-pix_fmt","rgb24"] if ext in("png","bmp") else []), p)
    return open(p, "rb").read()

def vid(ext, seed):
    n[0] += 1
    p = os.path.join(MEDIA, f"v{n[0]}.{ext}")
    extra = ["-c:v", "mjpeg"] if ext == "avi" else ["-c:v", "mpeg4"]
    ff("-f", "lavfi", "-i", f"color=c=0x{(seed*40503)&0xffffff:06x}:size=160x120:d=1:r=5", "-metadata", "creation_time=1970-01-01T00:00:00Z" if False else "comment=x", *extra, p)
    return open(p, "rb").read()

def sc(title, ts, lat=0.0, lon=0.0, desc="", fav=False, people=None, alt=0.0, exif_geo=None):
    d = {"title": title, "description": desc, "imageViews": "1",
         "creationTime": {"timestamp": str(ts + 1000), "formatted": ""},
         "photoTakenTime": {"timestamp": str(ts), "formatted": datetime.fromtimestamp(ts, timezone.utc).strftime("%b %d, %Y, %I:%M:%S %p UTC")},
         "geoData": {"latitude": lat, "longitude": lon, "altitude": alt, "latitudeSpan": 0.0, "longitudeSpan": 0.0},
         "geoDataExif": exif_geo or {"latitude": 0.0, "longitude": 0.0, "altitude": 0.0},
         "url": "https://photos.google.com/photo/x", "googlePhotosOrigin": {"mobileUpload": {"deviceType": "ANDROID_PHONE"}}}
    if fav: d["favorited"] = True
    if people: d["people"] = [{"name": p} for p in people]
    return json.dumps(d, ensure_ascii=False, indent=2).encode()

def ts(s):  # "2019-06-15 07:10:10" UTC
    return int(datetime.strptime(s, "%Y-%m-%d %H:%M:%S").replace(tzinfo=timezone.utc).timestamp())

GP = "Takeout/Google Photos/"
A = []  # (path, bytes) for export A
J1 = img("jpg", 1)
A += [(GP+"Photos from 2019/IMG_20190615_101010.jpg", J1),
      (GP+"Photos from 2019/IMG_20190615_101010.jpg.json", sc("IMG_20190615_101010.jpg", ts("2019-06-15 07:10:10"), 52.52, 13.405, desc="Отпуск\nна море", fav=True, people=["Вадим", "Anna"], alt=34.5)),
      (GP+"Германия 2017/IMG_20190615_101010.jpg", J1),
      (GP+"Германия 2017/IMG_20190615_101010.jpg.json", sc("IMG_20190615_101010.jpg", ts("2019-06-15 07:10:10"), 52.52, 13.405)),
      (GP+"Германия 2017/metadata.json", json.dumps({"title": "Германия-2017 (поездка)", "description": "", "access": "protected", "date": {"timestamp": "1500000000"}}, ensure_ascii=False).encode())]
# truncated json name (real case) — title is authoritative
S = img("jpg", 2)
A += [(GP+"Photos from 2021/Screenshot_2021-06-29-13-43-47-185_by.st.priorm.jpg", S),
      (GP+"Photos from 2021/Screenshot_2021-06-29-13-43-47-185_by.st.prior.json", sc("Screenshot_2021-06-29-13-43-47-185_by.st.priorm.jpg", ts("2021-06-29 10:43:47")))]
# truncated media name + truncated json, title has full name
L = img("jpg", 3)
A += [(GP+"Photos from 2021/IMG_9999_some_long_descriptive_filename_here_x.jpg", L),
      (GP+"Photos from 2021/IMG_9999_some_long_descriptive_filename_here_x.jp.json", sc("IMG_9999_some_long_descriptive_filename_here_x_full_original.jpg", ts("2021-01-02 03:04:05"), -33.8688, 151.2093))]
# (1) duplicates of name with different content
P1, P2 = img("jpg", 4), img("jpg", 5)
A += [(GP+"Photos from 2020/IMG_0001.jpg", P1), (GP+"Photos from 2020/IMG_0001.jpg.json", sc("IMG_0001.jpg", ts("2020-03-01 12:00:00"))),
      (GP+"Photos from 2020/IMG_0001(1).jpg", P2), (GP+"Photos from 2020/IMG_0001.jpg(1).json", sc("IMG_0001.jpg", ts("2020-04-02 12:00:00")))]
# edited
E0, E1 = img("jpg", 6), img("jpg", 7)
A += [(GP+"Photos from 2020/PXL_20200505_121212345.jpg", E0), (GP+"Photos from 2020/PXL_20200505_121212345.jpg.json", sc("PXL_20200505_121212345.jpg", ts("2020-05-05 09:12:12"), 55.75, 37.61)),
      (GP+"Photos from 2020/PXL_20200505_121212345-edited.jpg", E1)]
# live photo pair, one json
LP, LV = img("jpg", 8), vid("mp4", 8)
A += [(GP+"Photos from 2019/IMG_1234.JPG", LP), (GP+"Photos from 2019/IMG_1234.MP4", LV),
      (GP+"Photos from 2019/IMG_1234.JPG.json", sc("IMG_1234.JPG", ts("2019-08-01 15:00:00"), 53.849, 27.666))]
# no json, date in filename (WhatsApp)
A += [(GP+"Photos from 2019/IMG-20190101-WA0001.jpg", img("jpg", 9))]
# no json, no date -> year folder
A += [(GP+"Photos from 2019/random.png", img("png", 10))]
# video with geo
V = vid("mp4", 11)
A += [(GP+"Photos from 2019/VID_20190704_202020.mp4", V), (GP+"Photos from 2019/VID_20190704_202020.mp4.json", sc("VID_20190704_202020.mp4", ts("2019-07-04 17:20:20"), 41.9, 12.5, desc="Рим"))]
# AVI + THM companion
AV, TH = vid("avi", 12), img("jpg", 12)
A += [(GP+"Photos from 2018/MVI_0001.AVI", AV), (GP+"Photos from 2018/MVI_0001.THM", TH), (GP+"Photos from 2018/MVI_0001.AVI.json", sc("MVI_0001.AVI", ts("2018-02-03 10:00:00"), 50.45, 30.52))]
# gif
A += [(GP+"Photos from 2019/anim.gif", img("gif", 13)), (GP+"Photos from 2019/anim.gif.json", sc("anim.gif", ts("2019-09-09 09:09:09"), 1.5, 2.5))]
# Russian edited + Cyrillic file name in Cyrillic album "Дети"
R0, R1 = img("jpg", 14), img("jpg", 15)
A += [(GP+"Дети/фото.jpg", R0), (GP+"Дети/фото.jpg.json", sc("фото.jpg", ts("2016-12-31 21:30:00"))), (GP+"Дети/фото-изменено.jpg", R1)]
# existing EXIF date, no json
n[0] += 1; dp = os.path.join(MEDIA, "exif.jpg"); open(dp, "wb").write(img("jpg", 16))
subprocess.run(["exiftool", "-q", "-overwrite_original", "-DateTimeOriginal=2015:07:07 07:07:07", dp], check=True)
A += [(GP+"Дети/DSC_0001.JPG", open(dp, "rb").read())]
# existing EXIF date + json with different time -> keep existing tag, but GPS filled
n[0] += 1; dp2 = os.path.join(MEDIA, "exif2.jpg"); open(dp2, "wb").write(img("jpg", 17))
subprocess.run(["exiftool", "-q", "-overwrite_original", "-DateTimeOriginal=2014:04:04 04:04:04", dp2], check=True)
A += [(GP+"Дети/DSC_0002.JPG", open(dp2, "rb").read()), (GP+"Дети/DSC_0002.JPG.json", sc("DSC_0002.JPG", ts("2014-04-04 01:04:04"), 10.0, 20.0))]
# bmp undated, outside year folders
A += [(GP+"Дети/blob.bmp", img("bmp", 18))]
# Trash & Archive
A += [(GP+"Trash/deleted.jpg", img("jpg", 19)), (GP+"Trash/deleted.jpg.json", sc("deleted.jpg", ts("2017-01-01 10:00:00"))),
      (GP+"Archive/arch.jpg", img("jpg", 20)), (GP+"Archive/arch.jpg.json", sc("arch.jpg", ts("2017-02-02 10:00:00")))]
# json without media + top-level jsons
A += [(GP+"Photos from 2019/ghost.jpg.json", sc("ghost.jpg", ts("2019-01-01 00:00:00"))),
      (GP+"print-subscriptions.json", b'{"x":1}'), (GP+"shared_album_comments.json", b"[]"), (GP+"user-generated-memory-titles.json", b"{}")]

# Export B (newer, supplemental-metadata, overlaps A)
B = [(GP+"Photos from 2019/IMG_20190615_101010.jpg", J1),
     (GP+"Photos from 2019/IMG_20190615_101010.jpg.supplemental-metadata.json", sc("IMG_20190615_101010.jpg", ts("2019-06-15 07:10:10"), 52.52, 13.405)),
     (GP+"Photos from 2022/IMG_20220101_000001.jpg", img("jpg", 21)),
     (GP+"Photos from 2022/IMG_20220101_000001.jpg.supplemental-metadata.json", sc("IMG_20220101_000001.jpg", ts("2022-01-01 08:00:01"), 48.85, 2.35)),
     (GP+"Photos from 2022/PXL_20220202_101010123.MP.jpg", img("jpg", 22)),
     (GP+"Photos from 2022/PXL_20220202_101010123.MP.jpg.supplemental-meta.json", sc("PXL_20220202_101010123.MP.jpg", ts("2022-02-02 10:10:10"))),
     (GP+"Photos from 2022/IMG_5555(1).jpg", img("jpg", 23)),
     (GP+"Photos from 2022/IMG_5555.jpg.supplemental-metadata(1).json", sc("IMG_5555.jpg", ts("2022-03-03 03:03:03"))),
     (GP+"Отпуск в Крыму/IMG_20220101_000001.jpg", img("jpg", 21)),
     (GP+"Отпуск в Крыму/IMG_20220101_000001.jpg.supplemental-metadata.json", sc("IMG_20220101_000001.jpg", ts("2022-01-01 08:00:01"), 48.85, 2.35))]

class CP866Info(zipfile.ZipInfo):
    __slots__ = ()
    def _encodeFilenameFlags(self):
        return self.filename.encode("cp866"), self.flag_bits & ~0x800

def write_zip(path, items, cp866=False):
    with zipfile.ZipFile(path, "w", zipfile.ZIP_DEFLATED) as z:
        for name, data in items:
            zi = (CP866Info if cp866 else zipfile.ZipInfo)(name, (2022, 2, 24, 20, 9, 21))
            zi.compress_type = zipfile.ZIP_DEFLATED
            z.writestr(zi, data)

expA = os.path.join(ROOT, "D", "Google Photo Export"); os.makedirs(expA)
# split A over 3 parts; album "Германия 2017" split across part 1 and 2
parts = [A[0:4], A[4:16], A[16:]]
for i, p in enumerate(parts, 1):
    write_zip(os.path.join(expA, f"takeout-20220224T200921Z-00{i}.zip"), p)
# partially extracted sibling folder: one full file, one truncated (incomplete extraction)
ext = os.path.join(expA, "takeout-20220224T200921Z-001", GP, "Photos from 2019"); os.makedirs(ext)
open(os.path.join(ext, "IMG_20190615_101010.jpg"), "wb").write(J1)
open(os.path.join(ext, "IMG_20190615_101010.jpg.json"), "wb").write(A[1][1])
ext3 = os.path.join(expA, "takeout-20220224T200921Z-003", GP, "Photos from 2019"); os.makedirs(ext3)
open(os.path.join(ext3, "VID_20190704_202020.mp4"), "wb").write(V[: len(V)//2])
# export B: second location, one zip with legacy cp866 names (no UTF-8 flag)
expB = os.path.join(ROOT, "E", "Takeout2023"); os.makedirs(expB)
write_zip(os.path.join(expB, "takeout-2023-001.zip"), B[:8])
write_zip(os.path.join(expB, "takeout-2023-002.zip"), B[8:], cp866=True)
shutil.rmtree(MEDIA)
print("fixture at", ROOT)
