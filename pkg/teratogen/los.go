package teratogen

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
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

func blocksSightTranformed(center geom.Pt2I, xV geom.Vec2I) bool {
	hexPt := center.Plus(geom.Vec2I{xV.X, (xV.Y - xV.X) / 2})
	return GetArea().BlocksSight(hexPt)
}

func untransform(xV geom.Vec2I) geom.Vec2I { return geom.Vec2I{xV.X, (xV.Y - xV.X) / 2} }

func (self *Los) DoLos(center geom.Pt2I) {
	const losRadius = 12

	//	blocks := func(vec geom.Vec2I) bool { return GetArea().BlocksSight(center.Plus(vec)) }
	blocks := func(vec geom.Vec2I) bool { return blocksSightTranformed(center, vec) }

	outOfRadius := func(vec geom.Vec2I) bool {
		return geom.HexDist(center, center.Plus(untransform(vec))) > losRadius
	}

	for vec := range geom.LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(untransform(vec)))
	}
}

func CanSeeTo(start, end geom.Pt2I) bool {
	dist := 0
	// TODO Customizable max sight range
	sightRange := 18
	for o := range iterable.Drop(geom.Line(start, end), 1).Iter() {
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
