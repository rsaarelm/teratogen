// gen_version.go
//
// Copyright (C) 2013 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"os/exec"
)

var releases = map[string]string{}

func version() string {
	c := exec.Command("git", "log", "--pretty=format:%h", "-1")
	b, err := c.CombinedOutput()
	result := string(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read version")
		os.Exit(1)
	}

	if name, ok := releases[result]; ok {
		result = name
	}
	return result
}

func main() {
	ver := version()

	f, err := os.Create("src/teratogen/app/version.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(f, `// Generated file, do not edit

package app

const Version = "%s"
`, ver)
	f.Close()
}
