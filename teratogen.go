package main

import fmt "fmt"
import time "time"

import tcod "tcod"
import fomalhaut "fomalhaut"

const tickerWidth = 80;


func eatPrefix(str string, length int) (result string) {
	if len(str) < length {
		result = ""
	} else {
		result = str[length:1 + len(str) - length];
	}
	return;
}

func padString(str string, minLength int) (result string) {
	result = str;
	for ; len(result) < minLength; {
		result += " ";
	}
	return;
}

func updateTicker(str string, lineLength int) string {
	return padString(eatPrefix(str, 1), lineLength);
}

func main() {
	area := fomalhaut.NewMapField2();
	area.Set(10, 10, "A");

	fmt.Printf("Testing MapField2.\n");
	test, found := area.Get(10, 10);
	fmt.Printf("%v %v\n", test, found);

	test, found = area.Get(11, 10);
	fmt.Printf("%v %v\n", test, found);

	tcod.Init(80, 50, "Teratogen");
	tcod.SetForeColor(tcod.MakeColor(255, 255, 0));
	tcod.PutChar(0, 0, 64, tcod.BkgndNone);
	tcod.PrintLeft(0, 2, tcod.BkgndNone, "Hello, world!");
	tcod.SetForeColor(tcod.MakeColor(255, 0, 0));
	tcod.PutChar(0, 0, 65, tcod.BkgndNone);
	tcod.Flush();
	x := 40;
	y := 20;

	tickerLine := "                                                                                Teratogen online. ";

	go func() {
		for {
			lettersAtTime := 1;
			time.Sleep(int64(200000000 * lettersAtTime));
			for x := 0; x <= lettersAtTime; x++ {
				tickerLine = updateTicker(tickerLine, tickerWidth);
			}
		}
	}();

	tcod.SetForeColor(tcod.MakeColor(0, 255, 0));
	for {
		tcod.Clear();
		tcod.SetForeColor(tcod.MakeColor(192, 192, 192));
		tcod.PrintLeft(0, 0, tcod.BkgndNone, tickerLine);
		tcod.SetForeColor(tcod.MakeColor(0, 255, 0));

		tcod.PutChar(x, y, '@', tcod.BkgndNone);
		tcod.Flush();

		key := tcod.CheckForKeypress();
		switch key {
		case 'q':
			return;
		// Colemak direction pad.
		case 'n':
			x -= 1;
		case ',':
			y += 1;
		case 'i':
			x += 1;
		case 'u':
			y -= 1;
		case 'p':
			tickerLine += "Some text for the buffer... ";
		}
	}
}
