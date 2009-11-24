package fomalhaut

import "math"

type Vec2I [2]int32;

func (self Vec2I) Add(rhs Vec2I) {
	self[0] += rhs[0];
	self[1] += rhs[1];
}

func (self Vec2I) Subtract(rhs Vec2I) {
	self[0] -= rhs[0];
	self[1] -= rhs[1];
}

func (self Vec2I) ElemMultiply(rhs Vec2I) {
	self[0] *= rhs[0];
	self[1] *= rhs[1];
}

func (self Vec2I) ElemDivide(rhs Vec2I) {
	self[0] /= rhs[0];
	self[1] /= rhs[1];
}

func (self Vec2I) Dot(rhs Vec2I) int32 {
	return self[0] * rhs[0] + self[1] * rhs[1];
}

func (self Vec2I) Abs() float64 {
	return math.Sqrt(float64(self.Dot(self)));
}