package txt

import (
	"hyades/num"
	"regexp"
	"strings"
)

func EatPrefix(str string, length int) (result string) {
	if len(str) < length {
		result = ""
	} else {
		result = str[length:len(str)]
	}
	return
}

func PadString(str string, minLength int) (result string) {
	result = str
	for len(result) < minLength {
		result += " "
	}
	return
}

func Capitalize(str string) (result string) {
	if len(str) == 0 {
		return ""
	}
	return strings.ToUpper(str[0:1]) + str[1:]
}

// EditDistance returns the edit or Levenshtein distance between two strings.
// The edit distance is the minimum number of additions, deletions or changes
// of a single character to change one string to another.
func EditDistance(str1, str2 string) int {
	d := make([][]int, len(str1)+1)
	for i := 0; i < len(d); i++ {
		d[i] = make([]int, len(str2)+1)
	}

	for i := 0; i <= len(str1); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(str2); j++ {
		d[0][j] = j
	}

	for j := 1; j <= len(str2); j++ {
		for i := 1; i <= len(str1); i++ {
			if str1[i-1] == str2[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				del, ins, subst := d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+1
				d[i][j] = num.Imin(del, num.Imin(ins, subst))
			}
		}
	}
	return d[len(str1)][len(str2)]
}

var indefiniteArticleRegexp = regexp.MustCompile("^([aeio]|un|ul)")

// GuessIndefiniteArticle guesses whether a noun should get "a" or "an" as its
// indefinite article. It returns the article as an uncapitalized string.
func GuessIndefiniteArticle(noun string) string {
	noun = strings.ToLower(noun)
	if indefiniteArticleRegexp.MatchString(noun) {
		return "an"
	}
	return "a"
}
