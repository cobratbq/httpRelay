package httprelay

// Fuzz is go-fuzz's fuzzing function. Though there is not much to fuzz.
func Fuzz(data []byte) int {
	var val = string(data)
	var dropHdrs = map[string]struct{}{}
	bad := processConnectionHdr(dropHdrs, val)
	if len(dropHdrs) > 0 {
		return 1
	}
	if len(bad) == 0 {
		return 1
	}
	fullHost(val)
	return 0
}
