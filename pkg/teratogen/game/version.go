package game

import (
	"strconv"
)

const Version = "002"

func versionInt() int {
	ver, err := strconv.Atoi(Version)
	if err != nil {
		panic("Non-numeric version string")
	}
	return ver
}
