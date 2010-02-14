package teratogen

import (
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/mem"
	"io"
)

// A temporary component to hold the old Entity objects, before they get split
// up to subcomponents.

const BlobComponent = entity.ComponentFamily("blob")

type BlobHandler struct {
	blobs map[entity.Id]*Blob
}

func (self *BlobHandler) Init() { self.blobs = make(map[entity.Id]*Blob) }

// id2Blob converts ids to blob components, useful for iterable.Map.
func id2Blob(id interface{}) interface{} { return GetBlobs().Get(id.(entity.Id)) }

type Blob struct {
	IconId string
	guid   entity.Id
	Name   string
	pos    geom.Pt2I
	Class  EntityClass

	prop     map[string]interface{}
	hideProp map[string]bool
}

func NewEntity(guid entity.Id) (result *Blob) {
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

func (self *Blob) GetGuid() entity.Id { return self.guid }

func (self *Blob) GetClass() EntityClass { return self.Class }

func (self *Blob) GetName() string { return self.Name }

func (self *Blob) String() string { return self.Name }

func (self *Blob) MoveAbs(pos geom.Pt2I) { self.pos = pos }

func (self *Blob) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Blob) Set(name string, value interface{}) *Blob {
	self.hideProp[name] = false, false
	// Normalize
	switch a := value.(type) {
	case entity.Id:
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
	return GetWorld().GetEntity(prop.(entity.Id)), true
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
	case entity.Id:
		mem.WriteFixed(out, guidProp)
		mem.WriteFixed(out, int64(a))
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
		return entity.Id(mem.ReadInt64(in))
	default:
		dbg.Die("Bad prop type id %v", typ)
	}
	// Shouldn't get here.
	return nil
}

func (self *Blob) Serialize(out io.Writer) {
	mem.WriteString(out, self.IconId)
	mem.WriteFixed(out, int64(self.guid))
	mem.WriteString(out, self.Name)
	mem.WriteFixed(out, int32(self.pos.X))
	mem.WriteFixed(out, int32(self.pos.Y))
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
	self.guid = entity.Id(mem.ReadInt64(in))
	self.Name = mem.ReadString(in)
	self.pos.X = int(mem.ReadInt32(in))
	self.pos.Y = int(mem.ReadInt32(in))
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
