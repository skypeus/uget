// Copyright 2013-2015 Docker, Inc.
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

package units

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// See: http://en.wikipedia.org/wiki/Binary_prefix
const (
	// Decimal

	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

type unitMap map[string]int64

var (
	decimalMap = unitMap{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	binaryMap  = unitMap{"k": KiB, "m": MiB, "g": GiB, "t": TiB, "p": PiB}
	sizeRegex  = regexp.MustCompile(`^(\d+)([kKmMgGtTpP])?[bB]?$`)
)

var decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var binaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

// HumanSize returns a human-readable approximation of a size using SI standard (eg. "44kB", "17MB")
func HumanSize(size float64) string {
	return join(intToString(float64(size), 1000.0, decimapAbbrs))
}

// BytesSize returns a base-2 approximation of a size using IEC standard (eg. "44KiB", "17MiB")
func BytesSize(size float64) string {
	return join(Bytes(size))
}

func join(s, unit string) string {
	return fmt.Sprintf("%s %s", s, unit)
}

// Bytes returns a base-2 approximation of a size using IEC standard (eg. "44KiB", "17MiB")
func Bytes(size float64) (string, string) {
	return intToString(size, 1024.0, binaryAbbrs)
}

func intToString(size, unit float64, _map []string) (string, string) {
	i := 0
	for size >= unit {
		size = size / unit
		i++
	}
	s := fmt.Sprintf("%.4g", size)
	// trailing zeroes will be omitted, so add them in order to have uniform length
	// this only leaves 1000 - 1023 in bytes size to stand out with length 4 but that's okay
	for len(s) < 5 && (len(s) != 4 || strings.ContainsRune(s, '.')) {
		if strings.ContainsRune(s, '.') {
			s += "0"
		} else {
			s += ".0"
		}
	}
	return s, _map[i]
}

// FromHumanSize returns an integer from a human-readable specification of a
// size using SI standard (eg. "44kB", "17MB")
func FromHumanSize(size string) (int64, error) {
	return parseSize(size, decimalMap)
}

// RAMInBytes parses a human-readable string representing an amount of RAM
// in bytes, kibibytes, mebibytes, gibibytes, or tebibytes and
// returns the number of bytes, or -1 if the string is unparseable.
// Units are case-insensitive, and the 'b' suffix is optional.
func RAMInBytes(size string) (int64, error) {
	return parseSize(size, binaryMap)
}

// Parses the human-readable size string into the amount it represents
func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	size, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[2])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= mul
	}

	return size, nil
}
