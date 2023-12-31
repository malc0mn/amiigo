package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/nfcptl"
	"sync"
)

// config holds all the active settings
type config struct {
	// vendor is the vendor alias of the USB device to connect to.
	vendor string
	// device is the product alias of the USB device to connect to.
	device string
	// cacheDir is the path to the directory where data will be cached. If the path does not start
	// with a leading forward slash ("/"), it will be stored in the current users home directory.
	// It defaults to "~/.cache".
	cacheDir string
	// logFile is file path to write logs to. When set to an empty string, logs will be discarded.
	logFile string
	// amiiboApiBaseUrl is the base url for the open amiibo API by n3evin.
	amiiboApiBaseUrl string
	// retailKeyPath is the full path to a file containing concatenated unfixed-info.bin and
	// locked-secret.bin files.
	retailKeyPath string
	// retailKey is the loaded instance of the file referenced in retailKeyPath
	retailKey *amiibo.RetailKey
	// expertMode allows i.a. dangerous writes to NFC tokens that can cause defunct amiibo chars.
	// The token itself is not in danger!
	expertMode bool

	// ui holds UI related config.
	ui *uiConf

	// quit is closed on command shutdown.
	quit chan struct{}
	// wg is used for a clean shutdown.
	wg sync.WaitGroup
}

type uiConf struct {
	// invertImage will render images inverted as if they were selected by the cursor.
	invertImage bool
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
	logFile:          defaultLogFile,
	amiiboApiBaseUrl: defaultAmiiboApiBaseUrl,
	ui:               &uiConf{},
	quit:             make(chan struct{}),
}

func loadConfig(cFile string, conf *config) error {
	f, err := ini.Load(cFile)
	if err != nil {
		return err
	}

	if i, err := f.GetSection(""); err == nil {
		if k, err := i.GetKey("cache_dir"); err == nil {
			conf.cacheDir = k.String()
		}
		if k, err := i.GetKey("log_file"); err == nil {
			conf.logFile = k.String()
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
		if k, err := i.GetKey("retail_key"); err == nil {
			conf.retailKeyPath = k.String()
		}
	}

	if i, err := f.GetSection("ui"); err == nil {
		if k, err := i.GetKey("solid_images"); err == nil {
			if v, err := k.Bool(); err == nil {
				conf.ui.invertImage = v
			}
		}
	}

	return nil
}

func loadRetailKey(path string) (*amiibo.RetailKey, error) {
	if path == "" {
		return nil, nil
	}

	key, err := amiibo.NewRetailKey(path)
	if err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("config: retail key file %q is invalid: %s", path, err)
}
