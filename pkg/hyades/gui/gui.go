// Graphical user interface elements.

package gui

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/gfx"
)

type Font interface {
	Height() int
	LineWidth(txt string) int
}

type Widget interface {
	// Draw the widget using the assigned graphics context.
	Draw(g gfx.Graphics, area draw.Rectangle)

	// Gets an iteration of the node's immediate and their areas as
	// Window objects. The areas of children are assumed to be contained
	// in the area of the parent.
	Children(area draw.Rectangle) iterable.Iterable
}

// Window is a combination of a widget and it's draw area rectangle.
type Window struct {
	Widget Widget
	Area   draw.Rectangle
}

func (self *Window) Children() iterable.Iterable {
	return self.Widget.Children(self.Area)
}

// A drawing function that can serve as a simple leaf widget.
type DrawFunc func(g gfx.Graphics, area draw.Rectangle)

func (self DrawFunc) Draw(g gfx.Graphics, area draw.Rectangle) {
	self(g, area)
}

func (self DrawFunc) Children(area draw.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}

// DrawChildren is a helper function to iterate and draw the immediate
// children of a widget. It is inteded to be called from the Draw code of the
// parent widget.
func DrawChildren(g gfx.Graphics, area draw.Rectangle, widget Widget) {
	for childWindow := range widget.Children(area).Iter() {
		child := childWindow.(*Window)
		child.Widget.Draw(g, child.Area)
	}
}

func iterInPoint(pos draw.Point, area draw.Rectangle, node Widget, c chan<- interface{}) {
	// XXX: Stupid temporary rect to check if pos is in area.
	if draw.Rect(pos.X, pos.Y, pos.X+1, pos.Y+1).In(area) {
		c <- &Window{node, area}
		for childWindow := range node.Children(area).Iter() {
			child := childWindow.(*Window)
			iterInPoint(pos, child.Area, child.Widget, c)
		}
	}
}

// WidgetsContaining returns widgets whose areas contain pos ordered from the
// topmost to bottommost in draw order.
func WidgetsContaining(pos draw.Point, area draw.Rectangle, root Widget) iterable.Iterable {
	return alg.ReverseIter(alg.IterFunc(func(c chan<- interface{}) {
		iterInPoint(pos, area, root, c)
		close(c)
	}))
}

type TickHandler interface {
	HandleTickEvent(elapsedNs int64)
}

func DispatchTickEvent(widget Widget, area draw.Rectangle, elapsedNs int64) {
	window := &Window{widget, area}
	if tickee, ok := window.Widget.(TickHandler); ok {
		tickee.HandleTickEvent(elapsedNs)
	}
	for o := range window.Children().Iter() {
		win := o.(*Window)
		DispatchTickEvent(win.Widget, win.Area, elapsedNs)
	}
}
