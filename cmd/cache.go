package main

import (
	"os"
	"path"
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

// createCacheDirs creates the base cache folder.
func createCacheDirs() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cache := path.Join(home, cacheDir)
	// 0755 here in case the ~/.cache folder doesn't exist yet.
	if err := os.MkdirAll(cache, 0755); err != nil {
		return err
	}

	// 0700 here to set the amiigo dir permissions.
	if err := os.Chmod(cache, 0700); err != nil {
		return err
	}

	// Set the global variable.
	cacheBase = cache

	// TODO: can we do this in a loop?
	full := path.Join(cacheBase, cacheImgDir)
	if err := os.Mkdir(full, 0700); err != nil {
		return err
	}
	cacheImages = full

	full = path.Join(cacheBase, cacheApiDir)
	if err := os.Mkdir(full, 0700); err != nil {
		return err
	}
	cacheApi = full
	// TODO: end

	return nil
}

// getFromImageCache returns os.FileInfo for the requested image file or nil when it does not exist.
func getFromImageCache(file string) os.FileInfo {
	return getFromCache(file, cacheImages)
}

// getFromApiCache returns os.FileInfo for the requested API file or nil when it does not exist.
func getFromApiCache(file string) os.FileInfo {
	return getFromCache(file, cacheApi)
}

// getFromCache returns os.FileInfo for the requested file in the specified dir or nil when it does
// not exist.
func getFromCache(file, dir string) os.FileInfo {
	full := path.Join(dir, file)
	if i, err := os.Stat(full); err != nil {
		return i
	}

	return nil
}
