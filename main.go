package main

import (
	"fmt"
	"os"
	path "path/filepath"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/ncpa0cpl/static-server/utils"
)

func main() {
	args := utils.ParseArgs(os.Args[1:], []string{
		"--help",
		"--aw",
		"--watch",
		"--auto-reload",
		"--nocache",
		"--noetag",
	})

	if args.NamedParams.Has("help") {
		fmt.Println("Usage: goserve [options] [directory]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --help              Print this help message.")
		fmt.Println("  --loglevel <level>  The log level. Default: info")
		fmt.Println("  --port <port>       The port to serve on. Default: 8080")
		fmt.Println("  --redirect <url>    Redirect all unmatched routes to a specified url.")
		fmt.Println("  --spa <filepath>    Specify a file to send for all unmatched routes.")
		fmt.Println("  --chunk-size <KB>   The size of chunks when streaming. Default: 500KB")
		fmt.Println("  --no-streaming      Disables the server ability to process Range requests and sending partial content.")
		fmt.Println("")
		fmt.Println("Hot Module Reload")
		fmt.Println("  --aw           Alias for '--watch --auto-reload'")
		fmt.Println("  --watch        When enabled, server will send fs events when files are changed. To listen to these add event listeners to `window.HMR` on client side.")
		fmt.Println("  --auto-reload  Automatically inject a script to html files that will reload the page on a 'watch' change event.")
		fmt.Println("")
		fmt.Println("Cache Headers Options")
		fmt.Println("  --maxage <seconds>   The max-age value to set in the Cache-Control header.")
		fmt.Println("  --nocache            Require browsers to re-validate etag on each resource load.")
		fmt.Println("  --noetag             Disable ETag generation.")
		fmt.Println("")
		fmt.Println("Server Cache")
		fmt.Println("  --cache:max <MB>     Maximum size of all files in the cache. Default: 100MB")
		fmt.Println("  --cache:flimit <MB>  Maximum size of single file that can be put in cache. Default: 10MB")
		return
	}

	if args.HasParam("spa") && args.HasParam("redirect") {
		fmt.Println("Cannot specify both --spa and --redirect.")
		return
	}

	var rootDir string
	if args.Input != "" {
		if path.IsAbs(args.Input) {
			rootDir = args.Input
		} else {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("Unable to determine the serve directory.")
				return
			}
			rootDir = path.Join(wd, args.Input)
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Unable to determine the serve directory.")
			return
		}
		rootDir = wd
	}

	server := echo.New()

	switch args.GetParam("loglevel", "info") {
	case "info":
		server.Logger.SetLevel(log.INFO)
	case "debug":
		server.Logger.SetLevel(log.DEBUG)
	case "warn":
		server.Logger.SetLevel(log.WARN)
	case "error":
		server.Logger.SetLevel(log.ERROR)
	default:
		server.Logger.SetLevel(log.OFF)
	}

	server.Logger.Info(fmt.Sprintf("Serving files from: %s", rootDir))

	AddFileRoutes(server, "", rootDir, &Configuration{
		RedirectTo:       args.GetParam("redirect", ""),
		SpaFile:          args.GetParam("spa", ""),
		ExcludeEtag:      args.NamedParams.Has("noetag"),
		MaxAge:           args.GetParamInt("maxage", 0),
		NoCache:          args.NamedParams.Has("nocache"),
		MaxCacheSize:     args.GetParamUint64("cache:max", 100),
		MaxCacheFileSize: args.GetParamUint64("cache:flimit", 10),
		Watcher:          args.NamedParams.Has("watch") || args.NamedParams.Has("aw"),
		AutoReload:       args.NamedParams.Has("auto-reload") || args.NamedParams.Has("aw"),
		ChunkSize:        args.GetParamUint64("chunk-size", 500)*1024,
		NoStreaming:      args.NamedParams.Has("no-streaming"),
	})

	port := args.GetParam("port", "8080")
	err := server.Start(fmt.Sprintf(":%s", port))
	server.Logger.Fatal(err)
}
