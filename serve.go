package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	. "github.com/ncpa0cpl/convenient-structures"
	"github.com/ncpa0cpl/static-server/utils"
)

type StaticFile struct {
	Path              string
	RelPath           string
	Content           []byte
	ContentType       string
	LastModifiedAt    *time.Time
	LastModifiedAtRFC string
	Etag              string
}

var staticFiles *Array[*StaticFile] = &Array[*StaticFile]{}

func detectContentType(filepath string, content []byte) string {
	httpDet := http.DetectContentType(content)
	ext := path.Ext(filepath)

	if ext == ".svg" {
		return "image/svg+xml; charset=utf-8"
	}

	if strings.HasPrefix(httpDet, "text/plain") {
		switch ext {
		case ".html":
			return strings.Replace(httpDet, "text/plain", "text/html", 1)
		case ".css":
			return strings.Replace(httpDet, "text/plain", "text/css", 1)
		case ".js":
			return strings.Replace(httpDet, "text/plain", "text/javascript", 1)
		case ".json":
			return strings.Replace(httpDet, "text/plain", "application/json", 1)
		case ".xml":
			return strings.Replace(httpDet, "text/plain", "application/xml", 1)
		}
	}

	return httpDet
}

func (f *StaticFile) Revalidate() error {
	// check if the file has changed since last time
	// and reload it if it has
	file, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	modTime := info.ModTime()
	if modTime.Equal(*f.LastModifiedAt) {
		return nil
	}

	buff := make([]byte, info.Size())
	_, err = file.Read(buff)

	if err != nil {
		return err
	}

	f.Content = buff
	f.Etag = utils.HashBytes(buff)
	f.LastModifiedAt = &modTime
	f.LastModifiedAtRFC = modTime.Format(http.TimeFormat)
	f.ContentType = detectContentType(f.Path, buff)

	return nil
}

func getStaticFile(filepath string) ([]byte, string, *time.Time, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, "", nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, "", nil, err
	}

	buff := make([]byte, info.Size())
	_, err = file.Read(buff)

	if err != nil {
		return nil, "", nil, err
	}

	modTime := info.ModTime()
	return buff, detectContentType(filepath, buff), &modTime, err
}

type StaticResponse struct {
	file                     *StaticFile
	cacheMaxAge              int
	cacheRequireRevalidation bool
	acceptRangeRequests      bool
	isPrivate                bool
	sendInstead              error
	shouldSendInstead        bool
	contentType              string
}

func (s *StaticResponse) GetFilepath() string {
	return s.file.Path
}

func (s *StaticResponse) GetFileContent() []byte {
	// return the copy of the byte slice to avoid problems
	// that could be caused by the user mutating the array
	buff := make([]byte, len(s.file.Content))
	copy(buff, s.file.Content)
	return buff
}

func (s *StaticResponse) GetContentType() string {
	return s.file.ContentType
}

func (s *StaticResponse) GetLastModifiedAt() string {
	return s.file.LastModifiedAtRFC
}

func (s *StaticResponse) SetCacheMaxAge(age int) {
	s.cacheMaxAge = age
}

func (s *StaticResponse) SetNoCache(noCache bool) {
	s.cacheRequireRevalidation = noCache
}

func (s *StaticResponse) SetAcceptRangeRequests(allow bool) {
	s.acceptRangeRequests = allow
}

func (s *StaticResponse) SetIsPrivate(isPrivate bool) {
	s.isPrivate = isPrivate
}

func (s *StaticResponse) SetContentType(contentType string) {
	s.contentType = contentType
}

func (s *StaticResponse) Instead(err error) {
	s.sendInstead = err
	s.shouldSendInstead = true
}

func (s *StaticResponse) buildCacheControlHeader(conf *Configuration) string {
	hvalue := ""

	if s.isPrivate {
		hvalue += "private"
	} else {
		hvalue += "public"
	}

	if s.cacheRequireRevalidation || conf.NoCache {
		hvalue += ", no-cache"
	} else if s.cacheMaxAge != 0 {
		var maxAge int
		if conf.MaxAge != 0 {
			maxAge = conf.MaxAge
		} else {
			maxAge = s.cacheMaxAge
		}
		hvalue += ", must-revalidate, max-age=" + strconv.Itoa(maxAge)
	}

	return hvalue
}

type Configuration struct {
	BeforeSend  func(*StaticResponse, echo.Context) error
	RedirectTo  string
	ExcludeEtag bool
	MaxAge      int
	NoCache     bool
}

func AddFileRoutes(server *echo.Echo, baseUrl string, rootDir string, conf *Configuration) {
	if rootDir[len(rootDir)-1] != '/' {
		rootDir += "/"
	}

	utils.Walk(rootDir, func(root string, dirs []string, files []string) error {
		for _, file := range files {
			filepath := path.Join(root, file)
			relativePath := filepath[len(rootDir):]
			content, ctype, modTime, err := getStaticFile(filepath)

			if err == nil {
				server.Logger.Debug(fmt.Sprintf("Adding file: %s", relativePath))
				staticFiles.Push(
					&StaticFile{
						Path:              filepath,
						RelPath:           relativePath,
						Content:           content,
						ContentType:       ctype,
						Etag:              utils.HashBytes(content),
						LastModifiedAt:    modTime,
						LastModifiedAtRFC: modTime.Format(http.TimeFormat),
					},
				)
			}
		}
		return nil
	})

	server.GET(baseUrl+"/*", func(c echo.Context) error {
		routePath := c.Param("*")

		server.Logger.Debug(fmt.Sprintf("Received request for file: %s", routePath))

		sfIterator := staticFiles.Iterator()
		for !sfIterator.Done() {
			file, _ := sfIterator.Next()
			if file.RelPath == routePath {
				file.Revalidate()
				return sendFile(file, c, conf)
			}
		}

		// check if files exists in fs, and if it does load it into memory
		// and serve it
		filepath := path.Join(rootDir, routePath)
		content, ctype, modTime, err := getStaticFile(filepath)

		if err == nil {
			file := &StaticFile{
				Path:              filepath,
				RelPath:           routePath,
				Content:           content,
				ContentType:       ctype,
				Etag:              utils.HashBytes(content),
				LastModifiedAt:    modTime,
				LastModifiedAtRFC: modTime.Format(http.TimeFormat),
			}
			staticFiles.Push(file)

			return sendFile(file, c, conf)
		}

		if conf.RedirectTo != "" {
			return c.Redirect(302, conf.RedirectTo)
		}

		return c.String(404, "Not found")
	})
}

func sendFile(file *StaticFile, c echo.Context, conf *Configuration) error {
	sresp := &StaticResponse{
		file:                     file,
		cacheMaxAge:              86400,
		cacheRequireRevalidation: false,
		acceptRangeRequests:      true,
		isPrivate:                false,
		contentType:              file.ContentType,
	}

	if conf.BeforeSend != nil {
		err := conf.BeforeSend(sresp, c)
		if err != nil {
			return err
		}
		if sresp.shouldSendInstead {
			return sresp.sendInstead
		}
	}

	if c.Request().Header.Get("If-None-Match") == file.Etag || c.Request().Header.Get("If-Modified-Since") == file.LastModifiedAtRFC {
		return c.NoContent(304)
	}

	h := c.Response().Header()
	h.Set("Last-Modified", file.LastModifiedAtRFC)
	h.Set("Date", time.Now().Format(http.TimeFormat))
	h.Set("Content-Type", sresp.contentType)
	h.Set("Cache-Control", sresp.buildCacheControlHeader(conf))

	if !conf.ExcludeEtag {
		h.Set("ETag", file.Etag)
	}

	if sresp.acceptRangeRequests {
		h.Set("Accept-Ranges", "bytes")
		requestedRange := utils.ParseRangeHeader(&h)
		if requestedRange != nil {
			contentLength := strconv.FormatInt(requestedRange.End-requestedRange.Start+1, 10)
			contentRange := ("bytes " +
				strconv.FormatInt(requestedRange.Start, 10) +
				"-" + strconv.FormatInt(requestedRange.End, 10) +
				"/" + strconv.FormatInt(int64(len(file.Content)), 10))
			h.Set("Content-Length", contentLength)
			h.Set("Content-Range", contentRange)

			return c.Blob(200, file.ContentType, file.Content[requestedRange.Start:requestedRange.End+1])
		}
	}

	return c.Blob(200, file.ContentType, file.Content)
}
