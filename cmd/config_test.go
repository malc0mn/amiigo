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
}

func TestLoadconfigOk(t *testing.T) {
	cFile = "testdata/test_ok.conf"
	loadConfig()

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
}

func TestLoadConfigWrongPath(t *testing.T) {
	cFile = "does-not-exist.conf"
	err := loadConfig()

	want := "open " + cFile + ": no such file or directory"
	if err == nil || err.Error() != want {
		t.Errorf("got %s; want %s", err, want)
	}
}
