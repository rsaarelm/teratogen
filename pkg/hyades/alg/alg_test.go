package alg

import (
	"testing"
	"testing/quick"
)

func cmpStableInt(a, b int) bool {
	c1, ok1 := Cmp(a, b)
	c2, ok2 := Cmp(b, a)
	if (c1 == 0) != (a == b) {
		return false
	}
	return c1 == -c2 && ok1 && ok2
}

func cmpStableString(a, b string) bool {
	c1, ok1 := Cmp(a, b)
	c2, ok2 := Cmp(b, a)
	if (c1 == 0) != (a == b) {
		return false
	}
	return c1 == -c2 && ok1 && ok2
}

func cmpStableArr(a1, a2, a3, b1, b2, b3 byte) bool {
	a := []byte{a1, a2, a3}
	b := []byte{b1, b2, b3}

	c1, ok1 := Cmp(a, b)
	c2, ok2 := Cmp(b, a)
	eq := true
	for i, _ := range a {
		if a[i] != b[i] {
			eq = false
			break
		}
	}
	if (c1 == 0) != eq {
		return false
	}
	return c1 == -c2 && ok1 && ok2
}

type cmpStruct struct {
	a int
	b string
}

func cmpStableStruct(a1, a2 int, b1, b2 string) bool {
	s1 := cmpStruct{a1, b1}
	s2 := cmpStruct{a2, b2}

	c1, ok1 := Cmp(s1, s2)
	c2, ok2 := Cmp(s2, s1)

	if (c1 == 0) != (s1.a == s2.a && s1.b == s2.b) {
		return false
	}
	return c1 == -c2 && ok1 && ok2
}

func TestCmp(t *testing.T) {
	if err := quick.Check(cmpStableInt, nil); err != nil {
		t.Error(err)
	}
	if err := quick.Check(cmpStableString, nil); err != nil {
		t.Error(err)
	}
	if err := quick.Check(cmpStableArr, nil); err != nil {
		t.Error(err)
	}
	if err := quick.Check(cmpStableStruct, nil); err != nil {
		t.Error(err)
	}
}
