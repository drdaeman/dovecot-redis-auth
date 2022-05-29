package main

import (
	"strconv"
)

type BitFlags uint64

func (b BitFlags) Has(flag BitFlags) bool {
	return b&flag != 0
}

func ParseBitFlags(value string) (BitFlags, error) {
	flags, err := strconv.ParseUint(value, 10, 64)
	return BitFlags(flags), err
}
