package utils

import (
	"strconv"
	"strings"
	"time"
)

func GetInt64Range(s string, splitter string, limit func(int64) bool) (res []int64) {
	if len(s) <= len(splitter) {
		return nil
	}

	pair := strings.Split(s, splitter)
	if len(pair) != 2 {
		return nil
	}

	// Don't Trim, Frontend Should Do It
	//pair[0] = strings.TrimSpace(pair[0])
	//pair[1] = strings.TrimSpace(pair[1])

	var err error

	res = make([]int64, 2)
	res[0], err = strconv.ParseInt(pair[0], 10, 64)
	if err != nil {
		return nil
	}
	res[1], err = strconv.ParseInt(pair[1], 10, 64)
	if err != nil {
		return nil
	}

	if !limit(res[0]) || !limit(res[1]) || res[0] > res[1] {
		return nil
	}

	return res
}

func GetTimeRange(s string, splitter string, limit func(time.Time) bool) (res []time.Time /* YYYYMMDD */) {
	if len(s) <= len(splitter) {
		return nil
	}

	pair := strings.Split(s, splitter)
	if len(pair) != 2 {
		return nil
	}

	// Don't Trim, Frontend Should Do It
	//pair[0] = strings.TrimSpace(pair[0])
	//pair[1] = strings.TrimSpace(pair[1])

	var err error

	res = make([]time.Time, 2)
	res[0], err = time.ParseInLocation("20060102150405", pair[0], time.Local)
	if err != nil {
		return nil
	}
	res[1], err = time.ParseInLocation("20060102150405", pair[1], time.Local)
	if err != nil {
		return nil
	}

	if !limit(res[0]) || !limit(res[1]) || res[0].After(res[1]) {
		return nil
	}

	return res
}
