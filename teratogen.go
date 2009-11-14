package main

import fmt "fmt"
import tcod "tcod"

func main() {
	fmt.Printf("hello, world!\n");
	tcod.Init(80, 50, "Teratogen");
	tcod.SetForeColor(tcod.MakeColor(255, 255, 0));
	tcod.PutChar(0, 0, 64, tcod.BkgndNone);
	tcod.PrintLeft(0, 2, tcod.BkgndNone, "Hello, world!");
	tcod.SetForeColor(tcod.MakeColor(255, 0, 0));
	tcod.PutChar(0, 0, 65, tcod.BkgndNone);
	tcod.Flush();
	for {
		if tcod.CheckForKeypress() != 0 {
			return;
		}
	}
}
