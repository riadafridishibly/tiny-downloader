# Tiny Downloader
Tiny Downloader downloads file specified in the url. Specify number of goroutine by `-n N` to concurrently download (Default: 8).

## How To Run

```sh
$ go build
$ ./tiny-downloader <url>
```

## Example

```sh
$ go build
$ ./tiny-downloader http://ipv4.download.thinkbroadband.com/20MB.zip
$ ./tiny-downloader -n 16 http://ipv4.download.thinkbroadband.com/20MB.zip
```

## TODO
- [x] Variable number of goroutine
- [x] Add downloaded size in progress bar
- [ ] Add resume support
- [ ] Intelligent file divide (maybe power of 2?)
- [ ] Use single file like "aria2" rather than creating multiple ".part" files