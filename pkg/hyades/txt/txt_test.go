package txt

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

func editTest(t *testing.T, s1, s2 string, expected int) {
	actual := EditDistance(s1, s2)
	if expected != actual {
		t.Errorf("Expected '%v', got '%v'", expected, actual)
	}
}

func TestEditDistance(t *testing.T) {
	editTest(t, "", "", 0)
	editTest(t, "a", "a", 0)
	editTest(t, "a", "", 1)
	editTest(t, "", "a", 1)
	editTest(t, "foo", "", 3)
	editTest(t, "foo", "foo", 0)
	editTest(t, "foo", "fo", 1)
	editTest(t, "foo", "oo", 1)
	editTest(t, "foo", "fooo", 1)
	editTest(t, "foo", "foxo", 1)
	editTest(t, "foo", "fox", 1)
	editTest(t, "foo", "fxx", 2)
	editTest(t, "foo", "gya", 3)
}

func TestIndefiniteArticle(t *testing.T) {
	data := map[string]string{
		"aardvark":    "an",
		"behelit":     "a",
		"ettin":       "an",
		"hero":        "a",
		"illithid":    "an",
		"owlbear":     "an",
		"qwghlmian":   "a",
		"ukulele":     "a",
		"ultramarine": "an",
		"undead":      "an",
		"xorn":        "a",
		"zebranky":    "a",
	}

	for noun, expected := range data {
		article := GuessIndefiniteArticle(noun)
		if article != expected {
			t.Errorf("Got '%s %s', expected '%s %s'", article, noun, expected, noun)
		}
	}
}
