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

type Cache struct {
	maxSize     uint64
	maxFileSize uint64

	files       *Array[*StaticFile]
	currentSize uint64
}

func (c *Cache) CalcSize() uint64 {
	size := uint64(0)
	iter := c.files.Iterator()
	for !iter.Done() {
		file, _ := iter.Next()
		size += uint64(len(file.Content))
	}
	c.currentSize = size
	return size
}

func (c *Cache) CalcSizeMb() uint64 {
	bytesLen := c.CalcSize()
	return bytesLen / 1024 / 1024
}

func (c *Cache) Push(file *StaticFile) bool {
	fsize := uint64(len(file.Content))
	if fsize > c.maxFileSize {
		return false
	}
	if c.currentSize+fsize > c.maxSize {
		return false
	}
	c.currentSize += fsize
	c.files.Push(file)
	return true
}

func (c *Cache) Iterator() Iterator[*StaticFile] {
	return c.files.Iterator()
}

func (c *Cache) IsWithinFileLimit(file *StaticFile) bool {
	return uint64(len(file.Content)) <= c.maxFileSize
}

var cache *Cache = &Cache{
	files: &Array[*StaticFile]{},
}

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

// Return true if the file has changed
func (f *StaticFile) Revalidate() (bool, error) {
	// check if the file has changed since last time
	// and reload it if it has
	file, err := os.Open(f.Path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	modTime := info.ModTime()
	if modTime.Equal(*f.LastModifiedAt) {
		return false, nil
	}

	buff := make([]byte, info.Size())
	_, err = file.Read(buff)

	if err != nil {
		return false, err
	}

	f.Content = buff
	f.Etag = utils.HashBytes(buff)
	f.LastModifiedAt = &modTime
	f.LastModifiedAtRFC = modTime.Format(http.TimeFormat)
	f.ContentType = detectContentType(f.Path, buff)

	return true, nil
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
	BeforeSend       func(*StaticResponse, echo.Context) error
	RedirectTo       string
	ExcludeEtag      bool
	MaxAge           int
	NoCache          bool
	MacCacheSize     uint64
	MaxCacheFileSize uint64
}

func fmtSize(size int) string {
	if size < 1024 {
		return strconv.Itoa(size) + "B"
	}
	if size < 1024*1024 {
		return strconv.Itoa(size/1024) + "KB"
	}
	return strconv.Itoa(size/1024/1024) + "MB"
}

func AddFileRoutes(server *echo.Echo, baseUrl string, rootDir string, conf *Configuration) {
	cache.maxSize = conf.MacCacheSize * 1024 * 1024         // MB * KB * B = B
	cache.maxFileSize = conf.MaxCacheFileSize * 1024 * 1024 // MB * KB * B = B

	if rootDir[len(rootDir)-1] != '/' {
		rootDir += "/"
	}

	if cache.maxSize > 0 {
		utils.Walk(rootDir, func(root string, dirs []string, files []string) error {
			for _, file := range files {
				filepath := path.Join(root, file)
				relativePath := filepath[len(rootDir):]
				content, ctype, modTime, err := getStaticFile(filepath)

				if err == nil {
					server.Logger.Debugf("Adding file to cache: %s", relativePath)

					file := &StaticFile{
						Path:              filepath,
						RelPath:           relativePath,
						Content:           content,
						ContentType:       ctype,
						Etag:              utils.HashBytes(content),
						LastModifiedAt:    modTime,
						LastModifiedAtRFC: modTime.Format(http.TimeFormat),
					}
					added := cache.Push(file)

					if !added {
						if cache.IsWithinFileLimit(file) {
							server.Logger.Debugf(
								"Cache mem limit reached when adding file: %s (%s)",
								relativePath,
								fmtSize(len(file.Content)),
							)
							// stop walking the directory
							return fmt.Errorf("unable to add file to cache")
						}
					}
				}
			}
			return nil
		})

		server.Logger.Debugf(
			"Current cache size: %dMB",
			cache.CalcSizeMb(),
		)
	}

	server.GET(baseUrl+"/*", func(c echo.Context) error {
		routePath := c.Param("*")

		server.Logger.Debugf("Received request for file: %s", routePath)

		iter := cache.Iterator()
		for !iter.Done() {
			file, _ := iter.Next()
			if file.RelPath == routePath {
				changed, err := file.Revalidate()
				if err != nil {
					server.Logger.Errorf(
						"Failed to revalidate file(%s): %s",
						file.Path, err.Error(),
					)
					return c.String(500, "Internal server error")
				}
				if changed {
					// update cache size
					cache.CalcSize()
				}
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
			cache.Push(file)

			return sendFile(file, c, conf)
		} else {
			server.Logger.Errorf("Failed to read the file(%s): %s", filepath, err.Error())
		}

		if conf.RedirectTo != "" {
			server.Logger.Debugf(
				"Requested file not found, redirecting to: %s",
				conf.RedirectTo,
			)
			return c.Redirect(302, conf.RedirectTo)
		}

		server.Logger.Debug("Requested file not found")
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
		c.Logger().Debug("Resource not modified, returning 304")
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
