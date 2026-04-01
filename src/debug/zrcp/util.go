package zrcp

import (
	"fmt"
	"strconv"
	"strings"
)

func parseUint64(s string) (uint64, error) {
	v := strings.TrimSpace(strings.ToUpper(s))
	base := 0
	if strings.HasSuffix(v, "H") {
		v = strings.TrimSuffix(v, "H")
		base = 16
	}
	a, e := strconv.ParseUint(v, base, 64)
	return a, e
}

func parseUint16(s string) (uint16, error) {
	v, e := parseUint64(s)
	if v > 65535 {
		return uint16(v), fmt.Errorf("too big uint16: %d", v)
	}
	return uint16(v), e
}

func toW(hi, lo byte) uint16 {
	return uint16(lo) | (uint16(hi) << 8)
}

func iifStr(iif1, iif2 bool) string {
	flags := []byte{'-', '-'}
	if iif1 {
		flags[0] = '1'
	}
	if iif2 {
		flags[1] = '2'
	}
	return string(flags)
}

func typToString(typ uint8) string {
	switch typ {
	case 0:
		return "D"
	case 1:
		return "R"
	case 2:
		return "W"
	case 3:
		return "R/W"
	default:
		return "x"
	}
}
