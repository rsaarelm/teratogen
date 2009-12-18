package dbg

import (
	"fmt"
	"runtime"
)

func PrintBacktrace() {
	fmt.Print("\nStack trace:\n")

	for i := 0; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			fmt.Printf("%s:%d\n", file, line)
		} else {
			break
		}
	}
}

func Die(format string, a ...) {
	panic(fmt.Sprintf(format, a))

	// fmt.Print(msg + "\n")

	// XXX: Seems to crash the runtime currently (2009-11-19), even though
	// tracker says this is fixed:
	// (http://code.google.com/p/go/issues/detail?id=176).
	// PrintBacktrace()
	// os.Exit(1)
}

func Assert(exp bool, format string, a ...) {
	if !exp {
		Die(format, a)
	}
}

func AssertNotNil(val interface{}, format string, a ...) {
	Assert(val != nil, format, a)
}

func AssertNil(val interface{}, format string, a ...) {
	Assert(val == nil, format, a)
}
