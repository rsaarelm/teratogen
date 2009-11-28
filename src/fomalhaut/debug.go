package fomalhaut

import "fmt"
// import "os"
import "runtime"

func PrintBacktrace() {
	fmt.Print("\nStack trace:\n");

	for i := 0;; i++ {
		_, file, line, ok := runtime.Caller(i);
		if ok {
			fmt.Printf("%s:%d\n", file, line);
		} else {
			break;
		}
	}
}

// TODO: Formatting. (Dief?)
func Die(msg string) {
	panic(msg);
	// fmt.Print(msg + "\n");

	// XXX: Seems to crash the runtime currently (2009-11-19), even though
	// tracker says this is fixed:
	// (http://code.google.com/p/go/issues/detail?id=176).
	// PrintBacktrace();
	// os.Exit(1);
}

func DieIfNil(val interface{}, name string) {
	if val == nil { Die("Illegal nil value: "+name); }
}