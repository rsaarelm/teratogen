package main

import (
	"exp/iterable"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/mem"
	"io"
)

// A temporary component to hold the old Entity objects, before they get split
// up to subcomponents.

type BlobHandler struct {
	blobs map[entity.Id]*Blob
}


type Blob struct {
	IconId    string
	guid      Guid
	Name      string
	pos       geom.Pt2I
	parentId  Guid
	siblingId Guid
	childId   Guid
	Class     EntityClass

	prop     map[string]interface{}
	hideProp map[string]bool
}

func NewEntity(guid Guid) (result *Blob) {
	result = new(Blob)
	result.prop = make(map[string]interface{})
	result.hideProp = make(map[string]bool)
	result.guid = guid
	return
}

func (self *Blob) GetPos() geom.Pt2I {
	parent := self.GetParent()
	if parent != nil {
		return parent.GetPos()
	}
	return self.pos
}

func (self *Blob) GetGuid() Guid { return self.guid }

func (self *Blob) GetClass() EntityClass { return self.Class }

func (self *Blob) GetName() string { return self.Name }

func (self *Blob) String() string { return self.Name }

func (self *Blob) MoveAbs(pos geom.Pt2I) { self.pos = pos }

func (self *Blob) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Blob) GetParent() *Blob { return GetWorld().GetEntity(self.parentId) }

func (self *Blob) SetParent(e *Blob) {
	if e != nil {
		self.parentId = e.GetGuid()
	} else {
		self.parentId = *new(Guid)
	}
}

func (self *Blob) GetChild() *Blob { return GetWorld().GetEntity(self.childId) }

func (self *Blob) SetChild(e *Blob) {
	if e != nil {
		self.childId = e.GetGuid()
	} else {
		self.childId = *new(Guid)
	}
}

// GetSibling return the next sibling of the entity, or nil if there are none.
func (self *Blob) GetSibling() *Blob { return GetWorld().GetEntity(self.siblingId) }

func (self *Blob) SetSibling(e *Blob) {
	if e != nil {
		self.siblingId = e.GetGuid()
	} else {
		self.siblingId = *new(Guid)
	}
}

func (self *Blob) iterateChildrenWalk(c chan<- interface{}, recurse bool) {
	node := self.GetChild()
	for node != nil {
		c <- node
		if recurse {

			node.iterateChildrenWalk(c, recurse)
		}
		node = node.GetSibling()
	}
}

func (self *Blob) iterateChildren(c chan<- interface{}, recurse bool) {
	self.iterateChildrenWalk(c, recurse)
	close(c)
}

type entityContentIterable struct {
	e         *Blob
	recursive bool
}

func (self *entityContentIterable) Iter() <-chan interface{} {
	c := make(chan interface{})
	go self.e.iterateChildren(c, self.recursive)
	return c
}

// RecursiveContents iterates through all children and grandchildren of the
// entity.
func (self *Blob) RecursiveContents() iterable.Iterable {
	return &entityContentIterable{self, true}
}

// Contents iterates through the children but not the grandchildren of the
// entity.
func (self *Blob) Contents() iterable.Iterable {
	return &entityContentIterable{self, false}
}

func (self *Blob) HasContents() bool { return self.GetChild() != nil }

func (self *Blob) InsertSelf(parent *Blob) {
	self.RemoveSelf()
	if parent.GetChild() != nil {
		self.siblingId = parent.GetChild().GetGuid()
	}
	parent.SetChild(self)
	self.parentId = parent.GetGuid()
}

func (self *Blob) RemoveSelf() {
	parent := self.GetParent()
	self.parentId = *new(Guid)
	if parent != nil {
		if parent.GetChild().GetGuid() == self.GetGuid() {
			// First child of parent, modify parent's child
			// reference.
			parent.SetChild(self.GetSibling())
		} else {
			// Part of a sibling list, modify elder siblings
			// sibling reference.
			node := parent.GetChild()
			for {
				if node.GetSibling() == nil {
					dbg.Die("RemoveSelf: Entity not found among its siblings.")

				}
				if node.GetSibling().GetGuid() == self.GetGuid() {
					node.SetSibling(self.GetSibling())
					break
				}
				node = node.GetSibling()
			}
		}
		// Move to where the parent is in the world.
		self.MoveAbs(parent.GetPos())
	}
	self.siblingId = *new(Guid)
}

func (self *Blob) Set(name string, value interface{}) *Blob {
	self.hideProp[name] = false, false
	// Normalize
	switch a := value.(type) {
	case Guid:
	case string:
	// Ok as it is
	case float64:
	// Ok.
	case int:
		value = float64(a)
	case float:
		value = float64(a)
		// TODO: Other basic numeric types
	default:
		dbg.Die("Unsupported value type %#v", value)
	}
	self.prop[name] = value
	return self
}

// Convenience method.
func (self *Blob) SetFlag(name string) *Blob { return self.Set(name, 1) }

func (self *Blob) Get(name string) interface{} {
	_, hidden := self.hideProp[name]
	if hidden {
		return nil
	}
	ret, ok := self.prop[name]
	if ok {
		return ret
	}
	parent := self.PropParent()
	if parent != nil {
		return parent.Get(name)
	}
	return nil
}

func (self *Blob) GetF(name string) float64 {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return prop.(float64)
}

func (self *Blob) GetI(name string) int {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return int(prop.(float64))
}

func (self *Blob) GetIOpt(name string) (val int, ok bool) {
	prop := self.Get(name)
	if prop == nil {
		return
	}
	return int(prop.(float64)), true
}

func (self *Blob) GetS(name string) string {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return prop.(string)
}


func (self *Blob) GetSOpt(name string) (val string, ok bool) {
	prop := self.Get(name)
	if prop == nil {
		return
	}
	return prop.(string), true
}

// GetGuidOpt returns the entity for the guid in the properties if the
// property is present. If the property isn't a guid or if the guid's object
// can't be retrieved, runtime error.
func (self *Blob) GetGuidOpt(name string) (obj *Blob, ok bool) {
	prop := self.Get(name)
	if prop == nil {
		return
	}
	return GetWorld().GetEntity(prop.(Guid)), true
}

func (self *Blob) Has(name string) bool { return self.Get(name) != nil }

func (self *Blob) Hide(name string) *Blob {
	self.hideProp[name] = true
	return self
}

func (self *Blob) Clear(name string) *Blob {
	self.hideProp[name] = false, false
	self.prop[name] = nil, false
	return self
}

func (self *Blob) PropParent() *Blob {
	// TODO
	return nil
}

const (
	numProp    byte = 1
	stringProp byte = 2
	guidProp   byte = 3
)

func saveProp(out io.Writer, prop interface{}) {
	switch a := prop.(type) {
	case float64:
		mem.WriteFixed(out, numProp)
		mem.WriteFixed(out, a)
	case string:
		mem.WriteFixed(out, stringProp)
		mem.WriteString(out, a)
	case Guid:
		mem.WriteFixed(out, guidProp)
		mem.WriteString(out, string(a))
	default:
		dbg.Die("Bad prop type %#v", prop)
	}
}

func loadProp(in io.Reader) interface{} {
	switch typ := mem.ReadByte(in); typ {
	case numProp:
		return mem.ReadFloat64(in)
	case stringProp:
		return mem.ReadString(in)
	case guidProp:
		return Guid(mem.ReadString(in))
	default:
		dbg.Die("Bad prop type id %v", typ)
	}
	// Shouldn't get here.
	return nil
}

func (self *Blob) Serialize(out io.Writer) {
	mem.WriteString(out, self.IconId)
	mem.WriteString(out, string(self.guid))
	mem.WriteString(out, self.Name)
	mem.WriteFixed(out, int32(self.pos.X))
	mem.WriteFixed(out, int32(self.pos.Y))
	mem.WriteString(out, string(self.parentId))
	mem.WriteString(out, string(self.siblingId))
	mem.WriteString(out, string(self.childId))
	mem.WriteFixed(out, int32(self.Class))

	mem.WriteFixed(out, int32(len(self.prop)))
	for name, val := range self.prop {
		mem.WriteString(out, name)
		saveProp(out, val)
	}

	mem.WriteFixed(out, int32(len(self.hideProp)))
	for name, _ := range self.hideProp {
		mem.WriteString(out, name)
	}
}

func (self *Blob) Deserialize(in io.Reader) {
	self.IconId = mem.ReadString(in)
	self.guid = Guid(mem.ReadString(in))
	self.Name = mem.ReadString(in)
	self.pos.X = int(mem.ReadInt32(in))
	self.pos.Y = int(mem.ReadInt32(in))
	self.parentId = Guid(mem.ReadString(in))
	self.siblingId = Guid(mem.ReadString(in))
	self.childId = Guid(mem.ReadString(in))
	self.Class = EntityClass(mem.ReadInt32(in))

	self.prop = make(map[string]interface{})
	self.hideProp = make(map[string]bool)

	for i, numProps := 0, int(mem.ReadInt32(in)); i < numProps; i++ {
		name, val := mem.ReadString(in), loadProp(in)
		self.prop[name] = val
	}
	for i, numHide := 0, int(mem.ReadInt32(in)); i < numHide; i++ {
		self.hideProp[mem.ReadString(in)] = true
	}
}
