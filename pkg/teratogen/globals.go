package teratogen

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/entity"
	"hyades/mem"
	"io"
)

const GlobalsComponent = entity.ComponentFamily("globals")

// The Globals component holds global values.
type Globals struct {
	PlayerId     entity.Id
	AreaId       entity.Id
	CurrentLevel int32
	CurrentTurn  int64
}

func (self *Globals) Serialize(out io.Writer) { mem.GobSerialize(out, self) }

func (self *Globals) Deserialize(in io.Reader) {
	mem.GobDeserialize(in, self)
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

func GetCurrentTurn() int64 { return GetGlobals().CurrentTurn }

func NextTurn() { GetGlobals().CurrentTurn++ }
