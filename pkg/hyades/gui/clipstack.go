package gui

import (
	"container/vector"
	"exp/draw"
)

type ClipContext interface {
	SetClipRect(clipRect draw.Rectangle)
	ClearClipRect()
}

// A stack of clip rectangles for a clipping context, each constraining the
// draw area further.
type ClipStack interface {
	// PushClip sets the clip rectangle to an intersection of the previous rectangles and clipRect.
	PushClip(clipRect draw.Rectangle)

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

func (self *clipStack) PushClip(clipRect draw.Rectangle) {
	if self.stack.Len() > 0 {
		clipRect = clipRect.Clip(self.stack.Last().(draw.Rectangle))
	}
	self.stack.Push(clipRect)
	self.context.SetClipRect(clipRect)
}

func (self *clipStack) PopClip() {
	if self.stack.Len() > 0 {
		self.stack.Pop()
	}

	if self.stack.Len() > 0 {
		self.context.SetClipRect(self.stack.Last().(draw.Rectangle))
	} else {
		self.context.ClearClipRect()
	}
}
