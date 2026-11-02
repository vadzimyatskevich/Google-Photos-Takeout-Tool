package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reDateTime = regexp.MustCompile(`(?:^|[^0-9])((?:19|20)\d{2})[-_.]?(\d{2})[-_.]?(\d{2})[-_ T.]{0,3}(\d{2})[-_.:h]?(\d{2})[-_.:m]?(\d{2})(?:[^0-9]|\d{3}|$)`)
	reDate     = regexp.MustCompile(`(?:^|[^0-9])((?:19|20)\d{2})[-_.]?(\d{2})[-_.]?(\d{2})(?:[^0-9]|$)`)
	reYearDir  = regexp.MustCompile(`(?i)^(?:photos from|фото за|фото с|фотографии за)\s+((?:19|20)\d{2})$`)
	reExifDate = regexp.MustCompile(`^(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(Z|[+-]\d{2}:\d{2})?`)
)

func atoi(s string) int { v, _ := strconv.Atoi(s); return v }

func validDate(y, mo, d, h, mi, s int) bool {
	return y >= 1990 && y <= time.Now().Year()+1 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 && h < 24 && mi < 60 && s < 60
}

// dateFromName parses dates like IMG_20210524_110114, Screenshot_2021-06-29-13-43-47,
// PXL_20210101_123456789, IMG-20190101-WA0001, 2019-01-01 12.00.00.
func dateFromName(name string, loc *time.Location) (time.Time, bool) {
	if m := reDateTime.FindStringSubmatch(name); m != nil {
		y, mo, d, h, mi, s := atoi(m[1]), atoi(m[2]), atoi(m[3]), atoi(m[4]), atoi(m[5]), atoi(m[6])
		if validDate(y, mo, d, h, mi, s) {
			return time.Date(y, time.Month(mo), d, h, mi, s, 0, loc), true
		}
	}
	if m := reDate.FindStringSubmatch(name); m != nil {
		y, mo, d := atoi(m[1]), atoi(m[2]), atoi(m[3])
		if validDate(y, mo, d, 0, 0, 0) {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, loc), true
		}
	}
	return time.Time{}, false
}

func yearFromDir(dir string) int {
	if m := reYearDir.FindStringSubmatch(strings.TrimSpace(dir)); m != nil {
		return atoi(m[1])
	}
	return 0
}

func isAutoYearDir(dir string) bool { return yearFromDir(dir) != 0 }

// parseExifDate parses "YYYY:MM:DD HH:MM:SS[+hh:mm]"; naive values are in loc.
func parseExifDate(s string, loc *time.Location) (time.Time, bool) {
	m := reExifDate.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return time.Time{}, false
	}
	y, mo, d, h, mi, sec := atoi(m[1]), atoi(m[2]), atoi(m[3]), atoi(m[4]), atoi(m[5]), atoi(m[6])
	if !validDate(y, mo, d, h, mi, sec) {
		return time.Time{}, false
	}
	l := loc
	if m[7] == "Z" {
		l = time.UTC
	} else if m[7] != "" {
		sign := 1
		if m[7][0] == '-' {
			sign = -1
		}
		l = time.FixedZone("", sign*(atoi(m[7][1:3])*3600+atoi(m[7][4:6])*60))
	}
	return time.Date(y, time.Month(mo), d, h, mi, sec, 0, l).In(loc), true
}

func parseTZ(s string) (*time.Location, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "local") {
		return time.Local, nil
	}
	if strings.EqualFold(s, "utc") {
		return time.UTC, nil
	}
	if len(s) == 6 && (s[0] == '+' || s[0] == '-') && s[3] == ':' {
		sign := 1
		if s[0] == '-' {
			sign = -1
		}
		return time.FixedZone(s, sign*(atoi(s[1:3])*3600+atoi(s[4:6])*60)), nil
	}
	return time.LoadLocation(s)
}
