package gui

import (
	"container/vector"
	"hyades/num"
	"image"
)

type ClipContext interface {
	SetClipRect(clipRect image.Rectangle)
	ClearClipRect()
}

// A stack of clip rectangles for a clipping context, each constraining the
// draw area further.
type ClipStack interface {
	// PushClip sets the clip rectangle to an intersection of the previous rectangles and clipRect.
	PushClip(clipRect image.Rectangle)

	// PopClip reverts the last PushClip. It does nothing on an unpushed ClipStack.
	PopClip()
}

type clipStack struct {
	context ClipContext
	stack   *vector.Vector
}

func NewClipStack(context ClipContext) (result ClipStack) {
	return &clipStack{context, new(vector.Vector)}
}

func (self *clipStack) PushClip(clipRect image.Rectangle) {
	if self.stack.Len() > 0 {
		last := self.stack.Last().(image.Rectangle)
		clipRect = clipRect.Canon()
		clipRect = image.Rect(
			num.Imax(last.Min.X, clipRect.Min.X),
			num.Imax(last.Min.Y, clipRect.Min.Y),
			num.Imin(last.Max.X, clipRect.Max.X),
			num.Imin(last.Max.Y, clipRect.Max.Y))
	}
	self.stack.Push(clipRect)
	self.context.SetClipRect(clipRect)
}

func (self *clipStack) PopClip() {
	if self.stack.Len() > 0 {
		self.stack.Pop()
	}

	if self.stack.Len() > 0 {
		self.context.SetClipRect(self.stack.Last().(image.Rectangle))
	} else {
		self.context.ClearClipRect()
	}
}
