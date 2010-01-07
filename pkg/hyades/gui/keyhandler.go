package gui

import (
	"container/vector"
)

type KeyHandler interface {
	HandleKey(keyCode int)
}

type KeyHandlerFunc func(keyCode int)

func (self KeyHandlerFunc) HandleKey(keyCode int) {
	self(keyCode)
}

type KeyHandlerStack struct {
	stack *vector.Vector
}

func NewKeyHandlerStack() KeyHandlerStack { return KeyHandlerStack{new(vector.Vector)} }

func (self KeyHandlerStack) Push(handler KeyHandler) {
	self.stack.Push(handler)
}

func (self KeyHandlerStack) Peek() KeyHandler { return self.stack.Last().(KeyHandler) }

func (self KeyHandlerStack) Pop() KeyHandler { return self.stack.Pop().(KeyHandler) }

func (self KeyHandlerStack) IsEmpty() bool { return self.stack.Len() == 0 }
