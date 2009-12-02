package fomalhaut

// Ternary expression replacement.
func IfElse(exp bool, a interface{}, b interface{}) interface{} {
	if exp { return a; }
	return b;
}