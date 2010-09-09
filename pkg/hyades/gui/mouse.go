package gui

import (
	"exp/draw"
	"image"
)

const LeftMouseButton = 1 << 0
const MiddleMouseButton = 1 << 1
const RightMouseButton = 1 << 2
const MouseWheelUp = 1 << 3
const MouseWheelDown = 1 << 4

// MouseListener is an interface widgets can implement to receive mouse
// events.
type MouseListener interface {
	// MouseEvent takes a mouse state and may respond somehow. Returns
	// whether the event was consumed or if it should be passed on to the
	// next widget.
	HandleMouseEvent(area image.Rectangle, event draw.MouseEvent) (consumed bool)

	// MouseExited notifies the listener that the mouse has exited its area.
	MouseExited(event draw.MouseEvent)
}

// DispactchMouseEvent looks for widgets at where the mouse cursor is and that
// can receive mouse events. Sends the mouse event to one, if found. Returns
// the receiver, if one was found, nil otherwise. The parameter
// previousReceiver can be set to point to the MouseListener that received the
// previous mouse event. If it's different than the currently found one, it
// will be notified that the mouse has exited its area.
func DispatchMouseEvent(area image.Rectangle, root Widget, event draw.MouseEvent, previousReceiver MouseListener) MouseListener {
	pos := image.Pt(event.Loc.X, event.Loc.Y)
	for childWindow := range WidgetsContaining(pos, area, root).Iter() {
		child := childWindow.(*Window)
		if mouseReceiver, ok := child.Widget.(MouseListener); ok {
			if mouseReceiver.HandleMouseEvent(child.Area, event) {
				if previousReceiver != nil && mouseReceiver != previousReceiver {
					previousReceiver.MouseExited(event)
				}

				return mouseReceiver
			}
		}
	}
	if previousReceiver != nil {
		previousReceiver.MouseExited(event)
	}
	return nil
}
