package amiibo

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func loadDummyAmiibo(t *testing.T) *Amiibo {
	dmp, err := NewAmiibo(readFile(t, testDummyNtag), nil)
	if err != nil {
		t.Fatalf("NewAmiibo failed: got %s, want nil", err)
	}

	return dmp
}

func dummyAmiibo() []byte {
	return []byte{
		0xac, 0x51, 0x2c, 0x88, 0x1e, 0xe6, 0x35, 0x2a, 0x53, 0xe5, 0xe3, 0x78, 0x1f, 0x9c, 0x26, 0x0b, 0xbf, 0x66,
		0x59, 0x86, 0xb6, 0x87, 0x76, 0xb9, 0x05, 0xcb, 0xfc, 0xbf, 0x11, 0x13, 0x90, 0x4a, 0x9f, 0x5f, 0x0c, 0x4f,
		0x34, 0x15, 0xe6, 0x5f, 0x8f, 0x36, 0x67, 0xbe, 0x3c, 0xa5, 0xbf, 0x91, 0xf3, 0x16, 0x63, 0x76, 0xeb, 0x4e,
		0x2d, 0x62, 0xc0, 0xdf, 0xda, 0x26, 0x27, 0x6f, 0x97, 0x36, 0xb4, 0x9b, 0x09, 0x3e, 0x5f, 0xc9, 0x47, 0xea,
		0x7a, 0xe0, 0xc2, 0xd5, 0x6a, 0x74, 0x3d, 0x4b, 0xc5, 0x63, 0x60, 0xee, 0x19, 0x93, 0x08, 0x72, 0xdf, 0x8c,
		0x0d, 0xce, 0x17, 0xd5, 0x00, 0xd3, 0x29, 0xf5, 0xa0, 0xb1, 0xe7, 0x70, 0x34, 0x01, 0xb3, 0x3f, 0x12, 0x5b,
		0x9c, 0x6b, 0x18, 0xab, 0xf8, 0x3d, 0xaf, 0x92, 0xee, 0x83, 0xe6, 0x71, 0xb1, 0x90, 0x26, 0xdc, 0x2e, 0x2d,
		0x0e, 0x31, 0xf1, 0xfa, 0xb3, 0xb7, 0xe5, 0xe6, 0x37, 0x8a, 0xb4, 0x29, 0xfa, 0xb5, 0xb5, 0x22, 0xc0, 0xf3,
		0x42, 0x12, 0x3a, 0xbd, 0xdf, 0xa9, 0x40, 0xdf, 0x97, 0x57, 0x0b, 0x6f, 0x30, 0xc5, 0xa6, 0x26, 0x65, 0xe8,
		0x55, 0xc8, 0x3e, 0xb7, 0x76, 0xe5, 0x48, 0xe3, 0xe7, 0xf2, 0x56, 0x5f, 0xa3, 0xf5, 0x38, 0xae, 0x87, 0xe4,
		0xf5, 0x91, 0x1e, 0x32, 0x66, 0xe1, 0x1a, 0x6e, 0xf1, 0x39, 0x23, 0x9b, 0xd7, 0xad, 0x60, 0x98, 0xb9, 0xa6,
		0x27, 0xf6, 0x67, 0x2c, 0x02, 0xe4, 0x2c, 0x16, 0xad, 0x70, 0x23, 0x43, 0xfb, 0x0c, 0xaa, 0xa8, 0x11, 0x12,
		0x7f, 0xe2, 0x91, 0xf6, 0x61, 0x9d, 0xd2, 0xd7, 0x52, 0xfa, 0x67, 0x0f, 0x70, 0xe0, 0xcf, 0x0d, 0x65, 0x44,
		0x7e, 0xbc, 0xff, 0xff, 0x16, 0x79, 0xe1, 0xc3, 0x6a, 0xe2, 0x76, 0x78, 0x57, 0x42, 0x7d, 0x7a, 0x7e, 0x7f,
		0xa8, 0xa8, 0xb3, 0x3b, 0xea, 0xad, 0x3d, 0x4d, 0x56, 0x21, 0x43, 0xf1, 0xd3, 0x3e, 0x90, 0x16, 0x26, 0xda,
		0x62, 0xc9, 0x54, 0x00, 0xaa, 0x71, 0x83, 0x66, 0xd4, 0x3b, 0xe2, 0xdc, 0x1c, 0xd3, 0xff, 0x59, 0x7f, 0xf5,
		0x28, 0x99, 0xc0, 0xfc, 0x8e, 0x9f, 0x7f, 0xce, 0x28, 0x9b, 0xbc, 0x01, 0x9c, 0x78, 0xc7, 0xde, 0xf1, 0x25,
		0xa4, 0x6f, 0x6c, 0x39, 0x05, 0xaf, 0x2f, 0x8b, 0xe9, 0x10, 0xe7, 0x92, 0x94, 0x48, 0xdd, 0xbc, 0x59, 0x8c,
		0x3e, 0x4b, 0xb9, 0xe9, 0x99, 0x85, 0x79, 0x74, 0x70, 0xaa, 0x0d, 0x50, 0xb7, 0xb8, 0x7b, 0xc1, 0xc3, 0xad,
		0x94, 0x07, 0xd1, 0xd2, 0x34, 0xe2, 0x7e, 0xed, 0xa8, 0x57, 0xc1, 0x5f, 0x42, 0x86, 0xb6, 0x94, 0x2f, 0x63,
		0xed, 0x73, 0xc3, 0x75, 0xd3, 0x39, 0x6b, 0x7e, 0xc2, 0x3c, 0xf2, 0xdd, 0xf3, 0xbe, 0x98, 0xea, 0xcc, 0xf6,
		0x8c, 0xe3, 0x7f, 0x7f, 0x8a, 0x17, 0x80, 0x1f, 0x9f, 0xe7, 0x0d, 0xcc, 0x72, 0xa7, 0x04, 0xfa, 0x58, 0xee,
		0xcb, 0x3e, 0x24, 0x42, 0xf0, 0xab, 0xcb, 0xe9, 0xe4, 0xe8, 0x31, 0x86, 0x96, 0x60, 0xae, 0x1d, 0x61, 0x79,
		0x77, 0x20, 0x42, 0x21, 0xc7, 0x2d, 0x81, 0xc6, 0xd6, 0xd6, 0x74, 0xae, 0xb4, 0xf8, 0x30, 0x5e, 0xb6, 0xba,
		0x5d, 0x56, 0x7d, 0x11, 0x3f, 0x79, 0x93, 0x84, 0xcb, 0xd9, 0x1e, 0x3c, 0x8c, 0xe2, 0x0c, 0xb5, 0x05, 0xed,
		0xfb, 0x39, 0xf1, 0x7f, 0x2b, 0x69, 0x7b, 0x7a, 0xa4, 0x44, 0x6a, 0x37, 0x09, 0x13, 0x23, 0x4c, 0x45, 0x77,
		0xa8, 0xbe, 0xb7, 0xbc, 0xc4, 0x11, 0xfa, 0x1d, 0x37, 0x71, 0x0d, 0x36, 0x52, 0x40, 0x93, 0x5d, 0x4f, 0x74,
		0x7d, 0x78, 0x11, 0x62, 0x95, 0xa9, 0x13, 0x75, 0x26, 0x1d, 0x35, 0x21, 0x54, 0x27, 0xd7, 0x0e, 0x93, 0x18,
		0x77, 0x85, 0x7e, 0x95, 0x12, 0xac, 0x50, 0x89, 0x51, 0x83, 0xd2, 0x2f, 0xf5, 0xed, 0xad, 0x66, 0x01, 0x00,
		0x0f, 0xbd, 0x00, 0x00, 0x00, 0x04, 0x5f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func TestAmiibo_NewAmiiboNil(t *testing.T) {
	dmp, err := NewAmiibo(nil, nil)

	if dmp != nil || err == nil {
		t.Errorf("NewAmiibo should have failed, got: %v, %s", dmp, err)
	}
}

func TestAmiibo_NewAmiiboFileShort(t *testing.T) {
	dmp, err := NewAmiibo(readFile(t, testDummyNtag)[0:513], nil)
	if dmp != nil || err == nil {
		t.Errorf("NewAmiibo should have failed, got: %v, %s", dmp, err)
	}
}

func TestAmiibo_NewAmiiboFileLong(t *testing.T) {
	data := readFile(t, testDummyNtag)
	data = append(data, 0x12, 0xac, 0x50)

	dmp, err := NewAmiibo(data, nil)
	if dmp != nil || err == nil {
		t.Errorf("NewAmiibo should have failed, got: %v, %s", dmp, err)
	}
}

func TestAmiibo_NewAmiiboFileOk(t *testing.T) {
	dmp, err := NewAmiibo(readFile(t, testDummyNtag)[0:522], nil)
	if err != nil {
		t.Fatalf("got %s, want nil", err)
	}

	got := dmp.Raw()
	want := dummyAmiibo()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_NewAmiiboFilePerfect(t *testing.T) {
	dmp, err := NewAmiibo(readFile(t, testDummyNtag), nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}

	got := dmp.Raw()
	want := dummyAmiibo()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_NewAmiiboFromAmiitool(t *testing.T) {
	amiitool, err := NewAmiitool(readFile(t, testDummyAmiitool), nil)
	if amiitool == nil || err != nil {
		t.Fatalf("NewAmiitool should not failed, got: %v, %s", amiitool, err)
	}

	dmp, err := NewAmiibo(nil, amiitool)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}

	if dmp == nil {
		t.Error("got nil, want Amiibo struct")
	}
}

func TestAmiibo_Type(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.Type()
	want := TypeAmiibo

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestAmiibo_Unknown1(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.Unknown1()
	want := dummyUnknown1Data()

	if got != want {
		t.Errorf("got %#02x, want %#02x", got, want)
	}
}

func TestAmiibo_WriteCounter(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.WriteCounter()
	want := dummyWriteCounterData()

	if !bytes.Equal(got, want) {
		t.Errorf("got %#02x, want %#02x", got, want)
	}
}

func TestAmiibo_Unknown2(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.Unknown2()
	want := dummyUnknown2Data()

	if got != want {
		t.Errorf("got %#02x, want %#02x", got, want)
	}
}

func TestAmiibo_RegisterInfoRaw(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.RegisterInfoRaw()
	want := dummyRegisterInfoData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_RegisterInfo(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.RegisterInfo().data[:]
	want := dummyRegisterInfoData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_SetRegisterInfo(t *testing.T) {
	want := randomBytes(32)
	amiibo := loadDummyAmiibo(t)
	amiibo.SetRegisterInfo(want)
	got := amiibo.RegisterInfoRaw()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_TagHMAC(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.TagHMAC()
	want := dummyTagHmacData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_SetTagHMAC(t *testing.T) {
	want := randomBytes(32)
	amiibo := loadDummyAmiibo(t)
	amiibo.SetTagHMAC(want)
	got := amiibo.TagHMAC()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_ModelInfoRaw(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.ModelInfoRaw()
	want := dummyModelInfoData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_ModelInfo(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.ModelInfo().data[:]
	want := dummyModelInfoData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_Salt(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.Salt()
	want := dummySaltData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_DataHMAC(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.DataHMAC()
	want := dummyDataHmacData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_SetDataHMAC(t *testing.T) {
	want := randomBytes(32)
	amiibo := loadDummyAmiibo(t)
	amiibo.SetDataHMAC(want)
	got := amiibo.DataHMAC()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_SettingsRaw(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.SettingsRaw()
	want := dummySettingsData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_Settings(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	got := amiibo.Settings().data[:]
	want := dummySettingsData()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_SetSettings(t *testing.T) {
	want := randomBytes(360)
	amiibo := loadDummyAmiibo(t)
	amiibo.SetSettings(want)
	got := amiibo.SettingsRaw()

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestAmiibo_GeneratePassword(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	amiibo.GeneratePassword()

	gotPw := amiibo.Password()
	wantPw := dummyPassword()
	if !bytes.Equal(gotPw, wantPw) {
		t.Errorf("got %#04x, want %#04x", gotPw, wantPw)
	}

	gotPwa := amiibo.PasswordAcknowledge()
	wantPwa := dummyPasswordAcknowledge()
	if !bytes.Equal(gotPwa, wantPwa) {
		t.Errorf("got %#02x, want %#02x", gotPwa, wantPwa)
	}
}

func TestAmiibo_ResetSecurity(t *testing.T) {
	amiibo := loadDummyAmiibo(t)
	amiibo.GeneratePassword()

	pwd := amiibo.Password()
	amiibo.ResetSecurity()
	got := amiibo.Password()
	if bytes.Equal(got, pwd) {
		t.Errorf("got %#04x, want %#04x", got, pwd)
	}
}
