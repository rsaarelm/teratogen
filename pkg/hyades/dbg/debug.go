package dbg

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type Clock struct {
	accumulated int64
	timepoint   int64
}

func (self *Clock) Start() { self.timepoint = time.Nanoseconds() }

func (self *Clock) Stop() {
	self.accumulated += time.Nanoseconds() - self.timepoint
	self.timepoint = 0
}

func (self *Clock) Running() bool { return self.timepoint != 0 }

func (self *Clock) Nanoseconds() int64 { return self.accumulated }

var clocks map[string]*Clock

func init() { clocks = make(map[string]*Clock) }

func StartClock(name string) {
	clock, ok := clocks[name]
	if !ok {
		clock = new(Clock)
		clocks[name] = clock
	}
	if clock.Running() {
		Warn("StartClock: Clock '%s' already started.", name)
	} else {
		clock.Start()
	}
}

func StopClock(name string) {
	clock, ok := clocks[name]
	if !ok {
		Warn("StopClock: Clock %s never started.", name)
		return
	}
	if !clock.Running() {
		Warn("StopClock: Clock '%s' already stopped.", name)
	} else {
		clock.Stop()
	}
}

func PrintClocks() {
	for name, clock := range (clocks) {
		fmt.Printf("%s: %f\n", name, float64(clock.Nanoseconds())/1e9)
	}
}

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

func Die(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a))

	// fmt.Print(msg + "\n")

	// XXX: Seems to crash the runtime currently (2009-11-19), even though
	// tracker says this is fixed:
	// (http://code.google.com/p/go/issues/detail?id=176).
	// PrintBacktrace()
	// os.Exit(1)
}

func Assert(exp bool, format string, a ...interface{}) {
	if !exp {
		Die(format, a)
	}
}

func AssertNotNil(val interface{}, format string, a ...interface{}) {
	Assert(val != nil, format, a)
}

func AssertNil(val interface{}, format string, a ...interface{}) {
	Assert(val == nil, format, a)
}

// Make a note of a problem that isn't fatal but is still nice to know.
func Warn(format string, a ...interface{}) {
	fmt.Println("Warning: " + fmt.Sprintf(format, a))
}

func AssertNoError(err os.Error) {
	if err != nil {
		Die("Unhandled error: " + err.String())
	}
}
