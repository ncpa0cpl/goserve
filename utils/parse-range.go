package utils

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Range struct {
	HasStart bool
	HasEnd   bool
	Start uint64
	End   uint64
}

func ParseRangeHeader(c echo.Context) *Range {
	headers := c.Request().Header
	header := headers.Get("Range")

	if len(header) == 0 || !strings.HasPrefix(header, "bytes=") {
		return nil
	}

	header = strings.TrimPrefix(header, "bytes=")

	rangeParts := strings.Split(header, "-")

	r := &Range{}

	if len(rangeParts) > 0 && rangeParts[0] != "" {
		startValue, err := strconv.ParseUint(rangeParts[0], 10, 64)

		if err == nil {
			r.Start = startValue
			r.HasStart = true
		} else {
			c.Logger().Error(err)
			return nil
		}
	}

	if len(rangeParts) > 1 && rangeParts[1] != "" {
		endValue, err := strconv.ParseUint(rangeParts[1], 10, 64)

		if err == nil {
			r.End = endValue
			r.HasEnd = true
		} else {
			c.Logger().Error(err)
			return nil
		}
	}

	return r
}
