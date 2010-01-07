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

func (self *KeyHandlerStack) Init() { self.stack = new(vector.Vector) }

func (self *KeyHandlerStack) PushKeyHandler(handler KeyHandler) {
	self.stack.Push(handler)
}

func (self *KeyHandlerStack) PeekKeyHandler() KeyHandler {
	return self.stack.Last().(KeyHandler)
}

func (self *KeyHandlerStack) PopKeyHandler() KeyHandler {
	return self.stack.Pop().(KeyHandler)
}

func (self *KeyHandlerStack) KeyHandlersEmpty() bool {
	return self.stack.Len() == 0
}
