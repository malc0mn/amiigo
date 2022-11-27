package main

// zeroed returns true when the byt array only contains zeros, false otherwise.
func zeroed(p []byte) bool {
	for _, v := range p {
		if v != 0 {
			return false
		}
	}

	return true
}
