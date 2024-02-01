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
	args := utils.ParseArgs(os.Args[1:])

	if args.NamedParams.Has("help") {
		fmt.Println("Usage: serve [options] [directory]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --help              Print this help message.")
		fmt.Println("  --loglevel <level>  The log level. Default: info")
		fmt.Println("  --maxage <seconds>  The max-age value to set in the Cache-Control header.")
		fmt.Println("  --nocache           Disable caching.")
		fmt.Println("  --noetag            Disable ETag generation.")
		fmt.Println("  --port <port>       The port to serve on. Default: 8080")
		fmt.Println("  --redirect <url>    Redirect all unmatched routes to a specified url.")
		fmt.Println("  --cache:max <MB>    Maximum size of all files in the cache. Default: 100MB")
		fmt.Println("  --cache:flimit <MB> Maximum size of single file that can be put in cache. Default: 10MB")
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
		ExcludeEtag:      args.NamedParams.Has("noetag"),
		MaxAge:           args.GetParamInt("maxage", 0),
		NoCache:          args.NamedParams.Has("nocache"),
		MacCacheSize:     args.GetParamUint64("cache:max", 100),
		MaxCacheFileSize: args.GetParamUint64("cache:flimit", 10),
	})

	err := server.Start(fmt.Sprintf(":%s", args.GetParam("port", "8080")))
	server.Logger.Fatal(err)
}
