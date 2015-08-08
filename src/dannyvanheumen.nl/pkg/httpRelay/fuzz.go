package httpRelay

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
