package main

import (
	"hyades/dbg"
	"hyades/geom"
	"hyades/mem"
	"io"
)

type Entity struct {
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

func NewEntity(guid Guid) (result *Entity) {
	result = new(Entity)
	result.prop = make(map[string]interface{})
	result.hideProp = make(map[string]bool)
	result.guid = guid
	return
}

func (self *Entity) GetPos() geom.Pt2I {
	parent := self.GetParent()
	if parent != nil {
		return parent.GetPos()
	}
	return self.pos
}

func (self *Entity) GetGuid() Guid { return self.guid }

func (self *Entity) GetClass() EntityClass { return self.Class }

func (self *Entity) GetName() string { return self.Name }

func (self *Entity) MoveAbs(pos geom.Pt2I) { self.pos = pos }

func (self *Entity) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Entity) GetParent() *Entity { return GetWorld().GetEntity(self.parentId) }

func (self *Entity) SetParent(e *Entity) {
	if e != nil {
		self.parentId = e.GetGuid()
	} else {
		self.parentId = *new(Guid)
	}
}

func (self *Entity) GetChild() *Entity { return GetWorld().GetEntity(self.childId) }

func (self *Entity) SetChild(e *Entity) {
	if e != nil {
		self.childId = e.GetGuid()
	} else {
		self.childId = *new(Guid)
	}
}

// GetSibling return the next sibling of the entity, or nil if there are none.
func (self *Entity) GetSibling() *Entity { return GetWorld().GetEntity(self.siblingId) }

func (self *Entity) SetSibling(e *Entity) {
	if e != nil {
		self.siblingId = e.GetGuid()
	} else {
		self.siblingId = *new(Guid)
	}
}

func (self *Entity) iterateChildren(c chan<- *Entity) {
	node := self.GetChild()
	for node != nil {
		c <- node
		for i := range node.Contents() {
			c <- i
		}
		node = node.GetSibling()
	}

	close(c)
}

func (self *Entity) Contents() <-chan *Entity {
	c := make(chan *Entity)
	go self.iterateChildren(c)
	return c
}

func (self *Entity) InsertSelf(parent *Entity) {
	self.RemoveSelf()
	if parent.GetChild() != nil {
		self.siblingId = parent.GetChild().GetGuid()
	}
	parent.SetChild(self)
	self.parentId = parent.GetGuid()
}

func (self *Entity) RemoveSelf() {
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

func (self *Entity) Set(name string, value interface{}) *Entity {
	self.hideProp[name] = false, false
	// Normalize
	switch a := value.(type) {
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
func (self *Entity) SetFlag(name string) *Entity {
	return self.Set(name, 1)
}

func (self *Entity) Get(name string) interface{} {
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

func (self *Entity) GetF(name string) float64 {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return prop.(float64)
}

func (self *Entity) GetI(name string) int {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return int(prop.(float64))
}

func (self *Entity) GetS(name string) string {
	prop := self.Get(name)
	dbg.AssertNotNil(prop, "GetInt: Property %v not set", name)
	return prop.(string)
}

func (self *Entity) Has(name string) bool { return self.Get(name) != nil }

func (self *Entity) Hide(name string) *Entity {
	self.hideProp[name] = true
	return self
}

func (self *Entity) Clear(name string) *Entity {
	self.hideProp[name] = false, false
	self.prop[name] = nil, false
	return self
}

func (self *Entity) PropParent() *Entity {
	// TODO
	return nil
}

const (
	numProp    byte = 1
	stringProp byte = 2
)

func saveProp(out io.Writer, prop interface{}) {
	switch a := prop.(type) {
	case float64:
		mem.WriteByte(out, numProp)
		mem.WriteFloat64(out, a)
	case string:
		mem.WriteByte(out, stringProp)
		mem.WriteString(out, a)
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
	default:
		dbg.Die("Bad prop type id %v", typ)
	}
	// Shouldn't get here.
	return nil
}

func (self *Entity) Serialize(out io.Writer) {
	mem.WriteString(out, self.IconId)
	mem.WriteString(out, string(self.guid))
	mem.WriteString(out, self.Name)
	mem.WriteInt32(out, int32(self.pos.X))
	mem.WriteInt32(out, int32(self.pos.Y))
	mem.WriteString(out, string(self.parentId))
	mem.WriteString(out, string(self.siblingId))
	mem.WriteString(out, string(self.childId))
	mem.WriteInt32(out, int32(self.Class))

	mem.WriteInt32(out, int32(len(self.prop)))
	for name, val := range self.prop {
		mem.WriteString(out, name)
		saveProp(out, val)
	}

	mem.WriteInt32(out, int32(len(self.hideProp)))
	for name, _ := range self.hideProp {
		mem.WriteString(out, name)
	}
}

func (self *Entity) Deserialize(in io.Reader) {
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
