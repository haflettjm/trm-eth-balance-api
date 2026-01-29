package util

import "regexp"

var ethAddrRe = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

func IsValidEthAddress(s string) bool {
	return ethAddrRe.MatchString(s)
}

