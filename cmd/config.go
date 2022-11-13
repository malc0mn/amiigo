package main

import (
	"github.com/go-ini/ini"
	"github.com/malc0mn/amiigo/nfcptl"
)

// config holds all the active settings
type config struct {
	// vendor is the vendor alias of the USB device to connect to
	vendor string
	// device is the product alias of the USB device to connect to
	device string
	// cacheDir is the path to the directory where data will be cached. If the path does not start
	// with a leading forward slash ("/"), it will be stored in the current users home directory.
	// It defaults to "~/.cache".
	cacheDir string
	// amiamiiboApiBaseUrl
	amiiboApiBaseUrl string
}

const (
	// defaultVendor is the default vendor alias to use for vendor ID lookup
	defaultVendor = nfcptl.VendorDatelElextronicsLtd
	// defaultDevice is the default device alias to use for vendor ID lookup
	defaultDevice = nfcptl.ProductPowerSavesForAmiibo
	// defaultAmiiboApiBaseUrl is the default base url of the open Amiibo HTTP API.
	defaultAmiiboApiBaseUrl = "https://www.amiiboapi.com"
)

var conf = &config{
	vendor:           defaultVendor,
	device:           defaultDevice,
	cacheDir:         cacheDir,
	amiiboApiBaseUrl: defaultAmiiboApiBaseUrl,
}

func loadConfig() error {
	f, err := ini.Load(cFile)
	if err != nil {
		return err
	}

	if i, err := f.GetSection(""); err == nil {
		if k, err := i.GetKey("cache_dir"); err == nil {
			conf.cacheDir = k.String()
		}
		if k, err := i.GetKey("vendor"); err == nil {
			conf.vendor = k.String()
		}
		if k, err := i.GetKey("device"); err == nil {
			conf.device = k.String()
		}
		if k, err := i.GetKey("amiibo_api_base_url"); err == nil {
			conf.amiiboApiBaseUrl = k.String()
		}
	}

	return nil
}
