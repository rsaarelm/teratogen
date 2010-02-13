package teratogen

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWorld(t *testing.T) {
	context := NewContext()
	context.InitGame()

	if context.GetPlayer() == nil {
		t.Errorf("Player not initialized.")
	}

	// Store a string representation of player state for later.
	playerState := fmt.Sprintf("%#v", context.GetPlayer())

	// Try saving the game.
	file := new(bytes.Buffer)
	context.Serialize(file)

	context.InitGame()
	// Munge stuff a bit.
	context.EnterLevel(2)

	context = LoadContext(bytes.NewBuffer(file.Bytes()))

	// See that at least some of the state is the same.
	if playerState != fmt.Sprintf("%#v", context.GetPlayer()) {
		t.Errorf("Player state was changed during save-load.")
	}
}
