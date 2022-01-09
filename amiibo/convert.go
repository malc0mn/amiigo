package amiibo

// AmiiboToAmiitool converts a full 540 byte NTAG215 dump to internal amiitool format.
func AmiiboToAmiitool(amiibo *Amiibo) *Amiitool {
	d := [NTAG215Size]byte{}

	d[0] = amiibo.BCC1()
	d[1] = amiibo.Int()
	copy(d[2:4], amiibo.StaticLockBytes())
	copy(d[4:8], amiibo.CapabilityContainer())
	copy(d[8:40], amiibo.DataHMAC())
	d[40] = amiibo.Unknown1()
	copy(d[41:43], amiibo.WriteCounter())
	d[43] = amiibo.Unknown2()
	copy(d[44:76], amiibo.RegisterInfoRaw())
	copy(d[76:436], amiibo.SettingsRaw())
	copy(d[436:468], amiibo.TagHMAC())
	copy(d[468:476], amiibo.FullUID()[0:8])
	copy(d[476:488], amiibo.ModelInfoRaw())
	copy(d[488:520], amiibo.Salt())
	copy(d[520:], amiibo.Raw()[520:]) // Leftover NTAG215 data.

	return &Amiitool{data: d}
}

// AmiitoolToAmiibo converts the internal amiitool format to a NTAG215 dump.
func AmiitoolToAmiibo(amiibo *Amiitool) *Amiibo {
	d := [NTAG215Size]byte{}

	copy(d[:8], amiibo.FullUID()[:8])
	d[8] = amiibo.BCC1()
	d[9] = amiibo.Int()
	copy(d[10:12], amiibo.StaticLockBytes())
	copy(d[12:16], amiibo.CapabilityContainer())
	d[16] = amiibo.Unknown1()
	copy(d[17:19], amiibo.WriteCounter())
	d[19] = amiibo.Unknown2()
	copy(d[20:52], amiibo.RegisterInfoRaw())
	copy(d[52:84], amiibo.TagHMAC())
	copy(d[84:96], amiibo.ModelInfoRaw())
	copy(d[96:128], amiibo.Salt())
	copy(d[128:160], amiibo.DataHMAC())
	copy(d[160:520], amiibo.SettingsRaw())
	copy(d[520:], amiibo.Raw()[520:]) // Leftover NTAG215 data.

	return &Amiibo{NTAG215{data: d}}
}
