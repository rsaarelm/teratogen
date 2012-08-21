// ser.go
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
	"io"
)

type Serializable interface {
	// Serialize is the catch-all method that handles both serialization and
	// deserialization by changing the behavior of the Archive object that
	// visits the pointers to the member fields.
	Serialize(a Archive) error
}

type Archive interface {
	// Visit tells the archive to serialize or deserialize, depending on
	// archive type, the given pointer values.
	Visit(value ...interface{})

	// Input returns the io.Reader for a deserializing archive and nil for a
	// serializing one.
	Input() io.Reader

	// Output returns the io.Writer for a serializing archive and nil for a
	// deserializing one.
	Output() io.Writer

	// TagPointer explicitly marks a pointer to a pointer value for the
	// archive. The target of the pointer will need to be serialized
	// separately (once), and on deserialization the pointed pointer will need
	// to be rewritten to whatever the new address of the thing ends up being.
	// This method may be deprecated in favor of just using Visit once Visit
	// becomes sufficiently smart to deal with pointers.
	TagPointer(ptr interface{})
}

type base struct {
	// processedObject contains objects processed by the archive indexed by
	// their pre-serialization pointer.
	processedObjects map[uintptr]interface{}

	// seenObjects contains pointers seen by the archive that have not yet
	// been necessarily processed.
	seenObjects map[uintptr]interface{}
}

func (b *base) nextUnprocessed() (p uintptr, ok bool) {
	for ptr, _ := range b.seenObjects {
		if _, ok := b.processedObjects[ptr]; !ok {
			return ptr, true
		}
	}
	return
}
