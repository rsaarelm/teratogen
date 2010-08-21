package main

import (
	"exp/eval"
	"os"
	"strings"
	game "teratogen"
)

func RunRepl() {
	terp := eval.NewWorld()

	GetMsg().WriteString("Welcome to the console. Press return without writing anything to exit.\n")
	for {
		txt := GetMsg().InputText("> ")
		if txt == "" {
			return
		}

		code, err := terp.Compile(txt)

		// XXX: Hack to patch in a terminating semicolon if the error says one
		// is needed.
		if needTerminatingSemicolon(err) {
			code, err = terp.Compile(txt + ";")
		}

		if err != nil {
			game.Msg("Syntax error: %v\n", err)
		} else {
			ret, err := code.Run()
			if err != nil {
				game.Msg("%v\n", err)
			} else {
				game.Msg("%v\n", ret)
			}
		}
	}
}

func needTerminatingSemicolon(err os.Error) bool {
	// Check if the error is one complaining about missing end semicolon
	const terminatingSemicolonMsg = "expected ';', found 'EOF'"
	return err != nil && strings.Index(err.String(), terminatingSemicolonMsg) != -1
}
