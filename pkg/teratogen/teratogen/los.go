package teratogen

import (
	"hyades/entity"
	"hyades/geom"
	"io"
)

type LosState byte

const (
	LosUnknown LosState = iota
	LosMapped
	LosSeen
)

const LosComponent = entity.ComponentFamily("los")

// Line-of-sight state component.
type Los struct {
	w, h int
	los  []LosState
}

func NewLos() (result *Los) {
	result = new(Los)
	result.los = make([]LosState, mapWidth*mapHeight)
	return
}

func (self *Los) Serialize(out io.Writer) { entity.GobSerialize(out, self) }

func (self *Los) Deserialize(in io.Reader) { entity.GobDeserialize(in, self) }

func (self *Los) ClearSight() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		if self.los[idx] == LosSeen {
			self.los[idx] = LosMapped
		}
	}
}

// Debug command that makes the entire map visible.
func (self *Los) WizardEye() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosSeen
	}
}

func (self *Los) ClearMapped() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosUnknown
	}
}

func (self *Los) MarkSeen(pos geom.Pt2I) {
	if GetArea().InArea(pos) {
		self.los[pos.X+pos.Y*mapWidth] = LosSeen
	}
}

func (self *Los) Get(pos geom.Pt2I) LosState {
	if GetArea().InArea(pos) {
		return self.los[pos.X+pos.Y*mapWidth]
	}
	return LosUnknown
}

func (self *Los) DoLos(center geom.Pt2I) {
	const losRadius = 12

	blocks := func(vec geom.Vec2I) bool { return GetArea().BlocksSight(center.Plus(vec)) }

	outOfRadius := func(vec geom.Vec2I) bool { return int(vec.Abs()) > losRadius }

	for pt := range geom.LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(pt))
	}
}
