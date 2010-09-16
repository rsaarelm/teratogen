package main

import (
	"exp/eval"
	"os"
	"strings"
	"sync"
	"teratogen/game"
)

var onceTerp sync.Once

var terp *eval.World

func initTerp() {
	terp = eval.NewWorld()

	t, v := eval.FuncFromNativeTyped(wrapNextLevel, game.NextLevel)
	terp.DefineConst("NextLevel", t, v)
}

func wrapNextLevel(t *eval.Thread, args []eval.Value, res []eval.Value) {
	game.NextLevel()
}

func RunRepl() {
	onceTerp.Do(initTerp)

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
				if ret != nil {
					game.Msg("%v\n", ret)
				}
			}
		}
	}
}

// needTerminatingSemicolon checks if the error message is non-nil and
// complaining about missing end semicolon.
func needTerminatingSemicolon(err os.Error) bool {
	// XXX: Hardcoded specific error message wording, this will break if the
	// eval library changes its error message.
	const terminatingSemicolonMsg = "expected ';', found 'EOF'"
	return err != nil && strings.Index(err.String(), terminatingSemicolonMsg) != -1
}
