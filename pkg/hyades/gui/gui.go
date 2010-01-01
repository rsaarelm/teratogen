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
	Blit(img image.Image, x, y int)
	DrawString(font Font, x, y int, color image.Color, format string, a ...)
	DefaultFont() Font
	Clear(color image.Color)
	DrawRect(rect draw.Rectangle, color image.Color)
	FillRect(rect draw.Rectangle, color image.Color)
	DrawLine(p1, p2 draw.Point, color image.Color)
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
	// []interface{} {area draw.Rectangle, child Widget}. May return nil
	// if the widget has no children. The areas of children are assumed to
	// be contained in the area of the parent.
	Children(area draw.Rectangle) iterable.Iterable

	// If the widget responds to mouse events, return its mouse listener
	// interface. Otherwise return nil.
	GetMouseListener() MouseListener
}

type MouseListener func(area draw.Rectangle, event draw.Mouse)

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
		children := node.Children(area)
		if children != nil {
			for pair := range children.Iter() {
				childArea, childNode := UnpackWidgetIteration(pair)
				iterInPoint(pos, childArea, childNode, c)
			}
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
