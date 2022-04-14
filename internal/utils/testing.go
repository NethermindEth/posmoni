package utils

import (
	"fmt"
)

func CheckErr(descr string, isErr bool, err error) error {
	l := err == nil && isErr
	r := err != nil && !isErr

	if l || r {
		return fmt.Errorf("%s failed, unexpected error value: %v", descr, err)
	}
	return nil
}
