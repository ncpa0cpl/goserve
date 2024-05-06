package main

import (
	_ "embed"
	"strings"
)

//go:embed mimetypes
var mimetypes string

var MimeTypesMap = make(map[string]string)

func init() {
	lines := strings.Split(mimetypes, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) >= 2 && parts[0] != "" {
			contentType := parts[0]
			for i := 1; i < len(parts); i++ {
				if parts[i] != "" {
					MimeTypesMap[parts[i]] = contentType
				}
			}
		}
	}

	mimetypes = ""
}