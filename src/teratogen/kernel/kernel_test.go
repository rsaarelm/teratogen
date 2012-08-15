// kernel_test.go
//
// Copyright (C) 2012 Risto Saarelma
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

package kernel

import (
	"testing"
)

type countdownActor int

func (c countdownActor) Update(k Kernel) {
	//println(c)
	if c > 0 {
		k.Add(c - 1)
	}
}

func TestKernel(t *testing.T) {
	k := New()
	k.Add(countdownActor(10))
	k.Run()
}
