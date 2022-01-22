# Amiigo is the _uber geek amiibo inspector_

## Packages

## amiibo
The `amiibo` package can be used independently to work with amiibodumps being a
classic NTAG215 raw dump or a decrypted amiitool bin file.
It can decrypt and encrypt both formats and inspect or modify the amiibo data.

## nfcptl
The `nfcptl` package handles communications with NFC portal devices over USB. it
depends on the `gousb` package and provides a Client struct that handles the
device connection and communications.
This package can be used fully independently.

### Supported devices
- Datel's *PowerSaves for Amiibo*

### Should be supported
- NaMiio *NFC Backup System*
- MaxLander

From all the information found online, these two devices are **identical** to
Datel's *PowerSaves for Amiibo* device (the NaMiio device is 100% compatible
with the original Datel software which NaMiio even documented on their site)
but due to the lack of access to the hardware and the fact that they're no
longer commercially available they remain untested.

Granted: MaxLander uses 1K MFC tags but AFAIK they should be able to handle
NTAG215 as well from the data gathered. 1K MFC tag support is not present yet.

**Anyone out there care to have a go and post a report?**

### Would be cool to also support
- All amiibo related devices from Datel
- N2Elite USB reader/writer

Ideally hardware access to these devices is needed. Alternatively a full
wireshark dump of **all** operations would also be helpful.
**Or just create a pull request yourself!**

## Credits
- https://github.com/socram8888/amiitool
- https://www.3dbrew.org