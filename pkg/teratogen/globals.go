package teratogen

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/entity"
	"io"
)

const GlobalsComponent = entity.ComponentFamily("globals")

// The Globals component holds global values.
type Globals struct {
	PlayerId     entity.Id
	AreaId       entity.Id
	CurrentLevel int32
}

func (self *Globals) Serialize(out io.Writer) { entity.GobSerialize(out, self) }

func (self *Globals) Deserialize(in io.Reader) {
	entity.GobDeserialize(in, self)
}

func (self *Globals) Add(guid entity.Id, component interface{}) {
}

func (self *Globals) Remove(guid entity.Id) {}

func (self *Globals) Get(guid entity.Id) interface{} {
	return nil
}

func (self *Globals) EntityComponents() iterable.Iterable {
	return alg.EmptyIter()
}

func GetCurrentLevel() int { return int(GetGlobals().CurrentLevel) }
