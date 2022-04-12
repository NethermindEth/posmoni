package utils

import "testing"

func CheckErr(t *testing.T, descr string, isErr bool, err error) bool {
	l := err == nil && isErr
	r := err != nil && !isErr

	if l || r {
		t.Errorf("%s failed: %v", descr, err)
		return false
	}
	return true
}
