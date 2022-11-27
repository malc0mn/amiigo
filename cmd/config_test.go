package main

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	want := "datel"
	if conf.vendor != want {
		t.Errorf("conf.vendor = %s; want %s", conf.vendor, want)
	}

	want = "ps4amiibo"
	if conf.device != want {
		t.Errorf("conf.device = %s; want %s", conf.device, want)
	}

	want = ".cache/amiigo"
	if conf.cacheDir != want {
		t.Errorf("conf.cacheDir = %s; want %s", conf.cacheDir, want)
	}

	want = ""
	if conf.logFile != want {
		t.Errorf("conf.logFile = %s; want %s", conf.logFile, want)
	}

	want = "https://www.amiiboapi.com"
	if conf.amiiboApiBaseUrl != want {
		t.Errorf("conf.amiiboApiBaseUrl = %s; want %s", conf.amiiboApiBaseUrl, want)
	}
}

func TestLoadconfigOk(t *testing.T) {
	loadConfig("testdata/test_ok.conf", conf)

	want := "testvendor"
	if conf.vendor != want {
		t.Errorf("conf.vendor = %s; want %s", conf.vendor, want)
	}

	want = "testdevice"
	if conf.device != want {
		t.Errorf("conf.device = %s; want %s", conf.device, want)
	}

	want = "/some/test/dir"
	if conf.cacheDir != want {
		t.Errorf("conf.cacheDir = %s; want %s", conf.cacheDir, want)
	}

	want = "testing.log"
	if conf.logFile != want {
		t.Errorf("conf.logFile = %s; want %s", conf.logFile, want)
	}

	want = "https://example.com/api"
	if conf.amiiboApiBaseUrl != want {
		t.Errorf("conf.amiiboApiBaseUrl = %s; want %s", conf.amiiboApiBaseUrl, want)
	}

	wantB := true
	if conf.ui.invertImage != wantB {
		t.Errorf("conf.ui.invertImage = %v; want %v", conf.ui.invertImage, wantB)
	}
}

func TestLoadConfigWrongPath(t *testing.T) {
	cFile := "does-not-exist.conf"
	err := loadConfig(cFile, conf)

	want := "open " + cFile + ": no such file or directory"
	if err == nil || err.Error() != want {
		t.Errorf("got %s; want %s", err, want)
	}
}
