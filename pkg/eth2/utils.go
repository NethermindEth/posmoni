package eth2

import "strconv"

func parseUint(s string) (uint, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	return uint(i), err
}
