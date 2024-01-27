package utils

import (
	"net/http"
	"strconv"
	"strings"
)

type Range struct {
	Start int64
	End   int64
}

func ParseRangeHeader(headers *http.Header) *Range {
	header := headers.Get("Range")

	if len(header) == 0 || !strings.HasPrefix(header, "bytes=") {
		return nil
	}

	header = strings.TrimPrefix(header, "bytes=")

	rangeParts := strings.Split(header, "-")

	r := &Range{}

	if len(rangeParts) > 0 {
		startValue, err := strconv.ParseInt(rangeParts[0], 10, 64)

		if err == nil {
			r.Start = startValue
		} else {
			return nil
		}
	}

	if len(rangeParts) > 1 {
		endValue, err := strconv.ParseInt(rangeParts[1], 10, 64)

		if err == nil {
			r.End = endValue
		} else {
			return nil
		}
	}

	return r
}
