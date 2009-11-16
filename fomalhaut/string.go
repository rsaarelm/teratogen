package fomalhaut

func EatPrefix(str string, length int) (result string) {
	if len(str) < length {
		result = ""
	} else {
		result = str[length:1 + len(str) - length];
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