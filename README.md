## A very simple http static file server

### Build
```bash
go build -o ./serve
# Copy the binary to a bin directory (assuming ~/.local/bin is in your PATH)
cp ./serve ~/.local/bin
```

### Usage
```
Usage: serve [options] [directory]

Options:
  --help             Print this help message.
  --loglevel <level> The log level. Default: info
  --maxage <seconds> The max-age value to set in the Cache-Control header.
  --nocache          Set the Cache-Control header to no-cache.
  --noetag           Disable ETag header generation.
  --port <port>      The port to serve on. Default: 8080
  --redirect <url>   Redirect all unmatched routes to a specified url.
```

To serve files from the `public` directory of the current directory on port 8000:
```bash
serve --port 8000 ./public
```