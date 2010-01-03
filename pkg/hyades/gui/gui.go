// Graphical user interface elements.

package gui

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/gfx"
	"image"
)

type Font interface {
	Height() int
	LineWidth(txt string) int
}

type Widget interface {
	// Draw the widget using the assigned graphics context.
	Draw(g gfx.Graphics, area draw.Rectangle)

	// Gets an iteration of the node's immediate and their areas, as
	// []interface{} {area draw.Rectangle, child Widget}. The areas of
	// children are assumed to be contained in the area of the parent.
	Children(area draw.Rectangle) iterable.Iterable
}

// A drawing function that can serve as a simple leaf widget.
type DrawFunc func(g gfx.Graphics, area draw.Rectangle)

func (self DrawFunc) Draw(g gfx.Graphics, area draw.Rectangle) {
	self(g, area)
}

func (self DrawFunc) Children(area draw.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}

// MouseListener is an interface widgets can implement to receive mouse
// events.
type MouseListener interface {
	// MouseEvent takes a mouse state and may respond somehow. Returns
	// whether the event was consumed or if it should be passed on to the
	// next widget.
	HandleMouseEvent(area draw.Rectangle, event draw.Mouse) (consumed bool)
}

// DrawChildren is a helper function to iterate and draw the immediate
// children of a widget. It is inteded to be called from the Draw code of the
// parent widget.
func DrawChildren(g gfx.Graphics, area draw.Rectangle, widget Widget) {
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

// DispactchMouseEvent looks for widgets at where the mouse cursor is and that
// can receive mouse events. Sends the mouse event to one, if found. Returns
// whether a suitable widget was found.
func DispatchMouseEvent(area draw.Rectangle, root Widget, event draw.Mouse) bool {
	pos := draw.Pt(event.X, event.Y)
	for pair := range WidgetsContaining(pos, area, root).Iter() {
		area, widget := UnpackWidgetIteration(pair)
		if mouseReceiver, ok := widget.(MouseListener); ok {
			if mouseReceiver.HandleMouseEvent(area, event) {
				return true
			}
		}
	}
	return false
}

type TranslateGraphics struct {
	Vec   draw.Point
	Inner gfx.Graphics
}

func (self *TranslateGraphics) Width() int { return self.Inner.Width() }

func (self *TranslateGraphics) Height() int { return self.Inner.Height() }

func (self *TranslateGraphics) At(x, y int) image.Color {
	return self.Inner.At(x-self.Vec.X, y-self.Vec.Y)
}

func (self *TranslateGraphics) ColorModel() image.ColorModel {
	return self.Inner.ColorModel()
}

func (self *TranslateGraphics) Set(x, y int, c image.Color) {
	self.Inner.Set(x-self.Vec.X, y-self.Vec.Y, c)
}

func (self *TranslateGraphics) Blit(img image.Image, x, y int) {
	self.Inner.Blit(img, x-self.Vec.X, y-self.Vec.Y)
}

func (self *TranslateGraphics) FillRect(rect draw.Rectangle, color image.Color) {
	self.Inner.FillRect(rect.Sub(self.Vec), color)
}

func (self *TranslateGraphics) Center(x, y int) {
	self.Vec = draw.Pt(x-self.Width()/2, y-self.Height()/2)
}
