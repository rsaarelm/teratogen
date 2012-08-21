// ser_test.go
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

package ser

import (
	"bytes"
	"testing"
)

type linky struct {
	other *linky
	val   int
}

func (l *linky) Serialize(a Archive) error {
	a.Visit(&l.val)
	a.TagPointer(&l.other)
	return nil
}

func TestCycleSer(t *testing.T) {
	Register((*linky)(nil))

	cycle := &linky{val: 2}
	other := &linky{other: cycle, val: 3}
	cycle.other = other

	out := bytes.NewBuffer(nil)
	err := Save(cycle, out)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewBuffer(out.Bytes())

	obj, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	cycle2 := obj.(*linky)

	if cycle2.other == nil || cycle2.other.other != cycle2 || cycle2.other == cycle2 {
		t.Error("Bad cycle restore")
	}
	if cycle2.val != 2 || cycle2.other.val != 3 {
		t.Error("Bad cycle value restore")
	}
}
