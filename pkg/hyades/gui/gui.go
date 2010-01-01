// Graphical user interface elements.

package gui

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	//	"hyades/sdl"
	"image"
)

type Graphics interface {
	draw.Image
	Blit(img image.Image, x, y int)
	FillRect(rect draw.Rectangle, color image.Color)
	SetClipRect(clipRect draw.Rectangle)
	ClearClipRect()
}

type Font interface {
	Height() int
	LineWidth(txt string) int
}

type Widget interface {
	// Draw the widget using the assigned graphics context.
	Draw(g Graphics, area draw.Rectangle)

	// Gets an iteration of the node's immediate and their areas, as
	// []interface{} {area draw.Rectangle, child Widget}. The areas of
	// children are assumed to be contained in the area of the parent.
	Children(area draw.Rectangle) iterable.Iterable

	// If the widget responds to mouse events, return its mouse listener
	// interface. Otherwise return nil.
	GetMouseListener() MouseListener
}

// A drawing function that can serve as a simple leaf widget.
type DrawFunc func(g Graphics, area draw.Rectangle)

func (self DrawFunc) Draw(g Graphics, area draw.Rectangle) {
	self(g, area)
}

func (self DrawFunc) Children(area draw.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}

func (self DrawFunc) GetMouseListener() MouseListener {
	return nil
}

type MouseListener interface {
	MouseEvent(area draw.Rectangle, event draw.Mouse)
}

type FuncMouseListener func(area draw.Rectangle, event draw.Mouse)

func (self FuncMouseListener) MouseEvent(area draw.Rectangle, event draw.Mouse) {
	self(area, event)
}

// DrawChildren is a helper function to iterate and draw the immediate
// children of a widget. It is inteded to be called from the Draw code of the
// parent widget.
func DrawChildren(g Graphics, area draw.Rectangle, widget Widget) {
	for pair := range widget.Children(area).Iter() {
		childArea, child := UnpackWidgetIteration(pair)
		child.Draw(g, childArea)
	}
}

func UnpackWidgetIteration(pair interface{}) (area draw.Rectangle, widget Widget) {
	array := pair.([]interface{})
	area, widget = array[0].(draw.Rectangle), array[1].(Widget)
	return
}

func PackWidgetIteration(area draw.Rectangle, widget Widget) interface{} {
	return []interface{}{area, widget}
}

func iterInPoint(pos draw.Point, area draw.Rectangle, node Widget, c chan<- interface{}) {
	// XXX: Stupid temporary rect to check if pos is in area.
	if draw.Rpt(pos, pos).In(area) {
		c <- PackWidgetIteration(area, node)
		for pair := range node.Children(area).Iter() {
			childArea, childNode := UnpackWidgetIteration(pair)
			iterInPoint(pos, childArea, childNode, c)
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
