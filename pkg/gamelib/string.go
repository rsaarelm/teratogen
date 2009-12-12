package gamelib

import (
	"strings"
)

func EatPrefix(str string, length int) (result string) {
	if len(str) < length {
		result = ""
	} else {
		result = str[length:len(str)];
	}
	return;
}

func PadString(str string, minLength int) (result string) {
	result = str;
	for ; len(result) < minLength; {
		result += " ";
	}
	return;
}

func Capitalize(str string) (result string) {
	return strings.ToUpper(str[0:1]) + str[1:];
}
