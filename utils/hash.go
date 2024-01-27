package utils

import (
	"hash/crc64"
	"strconv"
)

func Hash(s string) string {
	return HashBytes([]byte(s))
}

var crcTable = crc64.MakeTable(crc64.ECMA)

func HashBytes(b []byte) string {
	checksum := crc64.Checksum(b, crcTable)
	return strconv.FormatUint(checksum, 16)
}
