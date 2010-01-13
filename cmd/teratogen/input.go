package main

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

type HandlerStack struct {
	stack *vector.Vector
}

func NewHandlerStack() HandlerStack { return HandlerStack{new(vector.Vector)} }

func (self HandlerStack) Push(handler KeyHandler) {
	self.stack.Push(handler)
}

func (self HandlerStack) Peek() KeyHandler { return self.stack.Last().(KeyHandler) }

func (self HandlerStack) Pop() KeyHandler { return self.stack.Pop().(KeyHandler) }

func (self HandlerStack) IsEmpty() bool { return self.stack.Len() == 0 }
