package main

import ()

type Item struct {
	EntityBase
}

func (self *Item) IsObstacle() bool { return false }
