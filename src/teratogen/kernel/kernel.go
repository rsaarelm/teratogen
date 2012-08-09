// kernel.go
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
	"container/list"
)

type Kernel interface {
	// Add adds a new actor to the end of the kernel's run queue
	Add(a Actor)

	// Run calls update on all actors in the kernel's run queue until the
	// queue is empty. Actors may add themselves or other actors to the kernel
	// when they are being updated.
	Run()
}

type Actor interface {
	Update(k Kernel)
}

type basicKernel struct {
	queue *list.List
}

func New() (result Kernel) {
	return &basicKernel{list.New()}
}

func (k *basicKernel) Add(a Actor) {
	k.queue.PushBack(a)
}

func (k *basicKernel) Run() {
	for k.queue.Len() > 0 {
		front := k.queue.Front()
		k.queue.Remove(front)
		front.Value.(Actor).Update(k)
	}
}
