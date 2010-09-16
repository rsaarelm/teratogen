package game

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
)

type FovState byte

const (
	FovUnknown FovState = iota
	FovMapped
	FovSeen
)

const FovComponent = entity.ComponentFamily("los")

// Line-of-sight state component.
type Fov struct {
	w, h int
	los  []FovState
}

func NewFov() (result *Fov) {
	result = new(Fov)
	result.los = make([]FovState, mapWidth*mapHeight)
	return
}

func (self *Fov) ClearSight() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		if self.los[idx] == FovSeen {
			self.los[idx] = FovMapped
		}
	}
}

// Debug command that makes the entire map visible.
func (self *Fov) WizardEye() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = FovSeen
	}
}

func (self *Fov) ClearMapped() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = FovUnknown
	}
}

func (self *Fov) MarkSeen(pos geom.Pt2I) {
	if GetArea().InArea(pos) {
		self.los[pos.X+pos.Y*mapWidth] = FovSeen
	}
}

func (self *Fov) Get(pos geom.Pt2I) FovState {
	if GetArea().InArea(pos) {
		return self.los[pos.X+pos.Y*mapWidth]
	}
	return FovUnknown
}

func (self *Fov) DoFov(center geom.Pt2I) {
	const losRadius = 12

	blocks := func(vec geom.Vec2I) bool { return GetArea().BlocksSight(center.Plus(vec)) }

	outOfRadius := func(vec geom.Vec2I) bool { return geom.HexDist(center, center.Plus(vec)) > losRadius }

	for vec := range geom.FieldOfView(geom.HexSectors, blocks, outOfRadius) {
		self.MarkSeen(center.Plus(vec))
	}
}

func CanSeeTo(start, end geom.Pt2I) bool {
	dist := 0
	// TODO Customizable max sight range
	sightRange := 18
	for o := range iterable.Drop(geom.HexLine(start, end), 1).Iter() {
		if dist > sightRange {
			return false
		}
		dist++
		pt := o.(geom.Pt2I)
		// Can see to the final cell even if that cell does block further sight.
		if pt.Equals(end) {
			break
		}
		if GetArea().BlocksSight(pt) {
			return false
		}
	}
	return true
}
