package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	cacheDir    = ".cache/amiigo"
	cacheImgDir = "images"
	cacheApiDir = "api-data"
)

var (
	cacheBase   string
	cacheImages string
	cacheApi    string
)

// cachedTransport represents a cached HTTP transport layer. For GET or HEAD requests, it will
// check the local filesystem for a cached response based on the full url. If no cached response
// was found, it will execute the call and when the request was successful (200 OK and no other
// errors), it will store the response body on the local filesystem.
type cachedTransport struct {
	t http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (ct *cachedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sum := md5.Sum([]byte(req.Host + req.URL.RequestURI()))
	name := hex.EncodeToString(sum[:]) + ".json"
	canCache := req.Method == "GET" || req.Method == "HEAD"

	if canCache {
		if f := getFromApiCache(name); f != nil {
			size := int64(-1) // -1 = unknown
			i, err := f.Stat()
			if err != nil {
				size = i.Size()
			}

			// More correct might have been to use a 304 Not Modified here but that would move some specifics to the
			// apii package to handle the Location header and fetch from the cache which does not feel right at all.
			headers := http.Header{}
			headers.Add("X-Local-Cache", "HIT")
			return &http.Response{
				Status:        http.StatusText(http.StatusOK),
				StatusCode:    http.StatusOK,
				Body:          f,
				ContentLength: size,
				Uncompressed:  true,
				Header:        headers,
			}, nil
		}
	}

	res, err := ct.t.RoundTrip(req)

	if canCache && err == nil && res.StatusCode == http.StatusOK {
		// Shame we cannot reset the http.Response.Body to allow multiple reads, so we use an io.TeeReader and replace
		// the body with a new reader.
		body := &bytes.Buffer{}
		if f, _ := writeToApiCache(io.TeeReader(res.Body, body), name); f != nil { // No error handling here. When the next identical request is done, it might succeed without problems.
			f.Close()
		}
		res.Body = io.NopCloser(body)
	}

	return res, err
}

// newCachedHttpClient returns a new http.Client which will check the local filesystem cache to see
// if the response for the requested url has been cached already. If not, it will do an actual HTTP
// call.
func newCachedHttpClient() *http.Client {
	return &http.Client{
		Transport: &cachedTransport{t: http.DefaultTransport},
		Timeout:   time.Second * 5,
	}
}

// createCacheDirs creates the amiibo caching folders inside the given path. If the path does not
// start with a forward slash, then it will be ceated inside the homedir of the user executing the
// binary.
func createCacheDirs(cache string) error {
	if !strings.HasPrefix(cache, "/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cache = path.Join(home, cache)
	}

	// 0755 in case the cache folder doesn't exist yet.
	if err := os.MkdirAll(cache, 0755); err != nil {
		return err
	}

	// 0700 to set the amiigo dir permissions.
	if err := os.Chmod(cache, 0700); err != nil {
		return err
	}

	// Set the global variable.
	cacheBase = cache

	// TODO: can we do this in a loop?
	full := path.Join(cacheBase, cacheImgDir)
	if err := os.MkdirAll(full, 0700); err != nil {
		return err
	}
	cacheImages = full

	full = path.Join(cacheBase, cacheApiDir)
	if err := os.MkdirAll(full, 0700); err != nil {
		return err
	}
	cacheApi = full
	// TODO: end

	return nil
}

// getImage gets the image data for the given url. It will look on the local filesystem first
// and when the file does not exist, will download it from the given url using http.Get().
func getImage(url string) (image.Image, error) {
	var f *os.File
	var err error
	if f = getFromImageCache(path.Base(url)); f == nil {
		if f, err = downloadAndCacheImage(url); err != nil {
			return nil, err
		}
	}

	i, _, err := image.Decode(f)
	f.Close()

	return i, err
}

// downloadAndCacheImage downloads an image from a given url and caches the data on the local
// filesystem. It returns a read only os.File pointer to the image file.
func downloadAndCacheImage(url string) (*os.File, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: http code %d", r.StatusCode)
	}

	return writeToImageCache(r.Body, path.Base(url))
}

// getFromImageCache returns *os.File for the requested image file or nil when it does not exist.
func getFromImageCache(file string) *os.File {
	return getFromCache(file, cacheImages)
}

// getFromApiCache returns *os.File for the requested API file or nil when it does not exist.
func getFromApiCache(file string) *os.File {
	return getFromCache(file, cacheApi)
}

// getFromCache returns a pointer to an os.File for the requested file in the specified dir or nil
// when it does not exist.
func getFromCache(file, dir string) *os.File {
	full := path.Join(dir, file)
	if _, err := os.Stat(full); err == nil {
		if f, err := os.Open(full); err == nil {
			return f
		}
	}

	return nil
}

// writeToImageCache writes an image to the local image cache dir and returns a read only os.File
// pointer to the image file.
func writeToImageCache(r io.Reader, file string) (*os.File, error) {
	return writeToCache(r, file, cacheImages)
}

// writeToApiCache writes a .json to the local API cache dir and returns a read only os.File
// pointer to the .json file.
func writeToApiCache(r io.Reader, file string) (*os.File, error) {
	return writeToCache(r, file, cacheApi)
}

// writeToCache writes a file to the local cache dir and returns a read only os.File pointer to the
// new file.
func writeToCache(r io.Reader, file, dir string) (*os.File, error) {
	p := path.Join(dir, file)
	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		return nil, err
	}
	f.Close()
	// f.Seek(0,0) would also work, but then we don't have a read only file to return!
	f, _ = os.Open(p)

	return f, nil
}
