package teratogen

import (
	"strconv"
)

const Version = "000"

func versionInt() int {
	ver, err := strconv.Atoi(Version)
	if err != nil {
		panic("Non-numeric version string")
	}
	return ver
}
