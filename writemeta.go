package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

func hasAny(tags map[string]string, keys ...string) bool {
	for _, k := range keys {
		if _, ok := tags[k]; ok {
			return true
		}
	}
	return false
}

func hasSuffixKey(tags map[string]string, suf string) bool {
	for k := range tags {
		if strings.HasSuffix(k, suf) {
			return true
		}
	}
	return false
}

func ff(v float64) string { return fmt.Sprintf("%.7f", v) }

// writeMeta writes metadata into g.Path (or an XMP sidecar for formats that can't hold it).
func (r *Runner) writeMeta(et *ExifTool, g *Group, tags map[string]string) {
	r.writeMetaX(et, g, tags, false)
}

func (r *Runner) writeMetaX(et *ExifTool, g *Group, tags map[string]string, forceSidecar bool) {
	o := r.o
	ext := strings.ToLower(filepath.Ext(g.Path))
	kind := writableExt[ext]
	if forceSidecar {
		kind = ""
	}
	target := g.Path
	sidecar := false
	if kind == "" {
		if !o.XMPSidecar {
			return
		}
		kind, sidecar = "xmp", true
		target = g.Path + ".xmp"
		if exists(target) {
			return
		}
		tags = map[string]string{}
	}
	var args []string
	set := func(present bool, tag, val string) {
		if o.Overwrite || !present {
			args = append(args, "-"+tag+"="+val)
		}
	}
	sc := g.SC
	dateOK := g.DateSrc == "json" || g.DateSrc == "filename" || (g.DateSrc == "exif" && o.Overwrite)
	if dateOK {
		t := g.Date
		dt := t.Format("2006:01:02 15:04:05")
		off := t.Format("-07:00")
		switch kind {
		case "img":
			if o.Overwrite || !hasAny(tags, "EXIF:DateTimeOriginal") {
				args = append(args, "-EXIF:DateTimeOriginal="+dt, "-EXIF:OffsetTimeOriginal="+off)
			}
			if o.Overwrite || !hasAny(tags, "EXIF:CreateDate") {
				args = append(args, "-EXIF:CreateDate="+dt, "-EXIF:OffsetTimeDigitized="+off)
			}
			if o.Overwrite || !hasAny(tags, "EXIF:ModifyDate") {
				args = append(args, "-EXIF:ModifyDate="+dt, "-EXIF:OffsetTime="+off)
			}
		case "gif", "xmp":
			set(hasAny(tags, "XMP:DateTimeOriginal"), "XMP-exif:DateTimeOriginal", dt+off)
			set(hasAny(tags, "XMP:CreateDate"), "XMP-xmp:CreateDate", dt+off)
			set(hasAny(tags, "XMP:DateCreated"), "XMP-photoshop:DateCreated", dt+off)
		case "qt":
			u := t.UTC().Format("2006:01:02 15:04:05") + "+00:00"
			p := hasAny(tags, "QuickTime:CreateDate")
			for _, tg := range []string{"QuickTime:CreateDate", "QuickTime:ModifyDate", "QuickTime:TrackCreateDate", "QuickTime:TrackModifyDate", "QuickTime:MediaCreateDate", "QuickTime:MediaModifyDate"} {
				set(p, tg, u)
			}
			set(hasAny(tags, "QuickTime:CreationDate"), "Keys:CreationDate", dt+off)
		}
	}
	gps := false
	if sc != nil && sc.HasGeo && (o.Overwrite || !hasSuffixKey(tags, ":GPSLatitude")) {
		gps = true
		lat, lon, alt := sc.Lat, sc.Lon, sc.Alt
		switch kind {
		case "img":
			args = append(args, "-EXIF:GPSLatitude="+ff(math.Abs(lat)), "-EXIF:GPSLatitudeRef="+ns(lat, "N", "S"),
				"-EXIF:GPSLongitude="+ff(math.Abs(lon)), "-EXIF:GPSLongitudeRef="+ns(lon, "E", "W"))
			if alt != 0 {
				args = append(args, "-EXIF:GPSAltitude="+ff(math.Abs(alt)), "-EXIF:GPSAltitudeRef="+ns(alt, "0", "1"))
			}
		case "gif", "xmp":
			args = append(args, "-XMP-exif:GPSLatitude="+ff(lat), "-XMP-exif:GPSLongitude="+ff(lon))
			if alt != 0 {
				args = append(args, "-XMP-exif:GPSAltitude="+ff(math.Abs(alt)), "-XMP-exif:GPSAltitudeRef="+ns(alt, "0", "1"))
			}
		case "qt":
			args = append(args, fmt.Sprintf("-Keys:GPSCoordinates=%s, %s, %s", ff(lat), ff(lon), ff(alt)))
		}
	}
	desc := false
	if sc != nil && sc.Description != "" {
		d := sc.Description
		has := hasAny(tags, "EXIF:ImageDescription", "XMP:Description", "IPTC:Caption-Abstract", "QuickTime:Description")
		if o.Overwrite || !has {
			desc = true
			switch kind {
			case "img":
				args = append(args, "-EXIF:ImageDescription="+d, "-XMP-dc:Description="+d)
				if ext == ".jpg" || ext == ".jpeg" || ext == ".jpe" {
					args = append(args, "-IPTC:CodedCharacterSet=UTF8", "-IPTC:Caption-Abstract="+d)
				}
			case "gif", "xmp":
				args = append(args, "-XMP-dc:Description="+d)
			case "qt":
				args = append(args, "-Keys:Description="+d, "-XMP-dc:Description="+d)
			}
		}
	}
	if o.Rating && sc != nil && sc.Favorited {
		set(hasAny(tags, "XMP:Rating"), "XMP-xmp:Rating", "5")
	}
	var kw []string
	if o.AlbumKeywords {
		kw = append(kw, g.Albums...)
	}
	if o.People && sc != nil {
		for _, pn := range sc.People {
			kw = append(kw, pn)
			args = append(args, "-XMP-iptcExt:PersonInImage-="+pn, "-XMP-iptcExt:PersonInImage+="+pn)
		}
	}
	for _, k := range kw {
		args = append(args, "-XMP-dc:Subject-="+k, "-XMP-dc:Subject+="+k)
	}
	if len(args) == 0 {
		return
	}
	full := append([]string{"-overwrite_original", "-m", "-api", "QuickTimeUTC=1", "-api", "LargeFileSupport=1"}, args...)
	full = append(full, target)
	so, se, err := et.Run(full...)
	ok := err == nil && (strings.Contains(so, "1 image files updated") || strings.Contains(so, "1 image files created"))
	if !ok {
		if r.p.cancel.Load() {
			return
		}
		if strings.Contains(so, "1 image files unchanged") {
			return
		}
		msg := strings.TrimSpace(se)
		if err != nil {
			msg = err.Error()
		}
		r.fail(g, r.L.T("err.exifline", oneLine(msg)))
		if !sidecar && o.XMPSidecar {
			r.writeMetaX(et, g, nil, true)
		}
		return
	}
	if sidecar {
		r.stat(func(s *Stats) { s.XMPSidecars++ })
		if !g.Date.IsZero() && g.DateSrc != "year" {
			setFileTimes(target, g.Date)
		}
	}
	r.stat(func(s *Stats) {
		s.MetaWritten++
		if gps {
			s.GPSWritten++
		}
		if desc {
			s.DescWritten++
		}
	})
	g.GPS = gps
}

func ns(v float64, pos, neg string) string {
	if v < 0 {
		return neg
	}
	return pos
}

func oneLine(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", " | ")
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}
