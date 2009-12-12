package gamelib

import (
	"testing"
)

func strEqual(t *testing.T, actual, expected string) {
	if expected != actual {
		t.Errorf("Expected '%v', got '%v'", expected, actual)
	}
}

func TestEatPrefix(t *testing.T) {
	strEqual(t, EatPrefix("", 1), "")
	strEqual(t, EatPrefix("", 10), "")
	strEqual(t, EatPrefix("xyzzy", 1), "yzzy")
	strEqual(t, EatPrefix("xyzzy", 2), "zzy")
	strEqual(t, EatPrefix("xyzzy", 3), "zy")
	strEqual(t, EatPrefix("xyzzy", 10), "")
}

func TestPadString(t *testing.T) {
	strEqual(t, PadString("", 10), "          ")
	strEqual(t, PadString("foobarbazquux", 10), "foobarbazquux")
	strEqual(t, PadString("foobar", 10), "foobar    ")
}
