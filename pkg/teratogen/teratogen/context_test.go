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

	save1 := file.Bytes()
	context = LoadContext(bytes.NewBuffer(save1))

	// See that at least some of the state is the same.
	if playerState != fmt.Sprintf("%#v", context.GetPlayer()) {
		t.Errorf("Player state was changed during save-load.")
	}

	file = new(bytes.Buffer)
	context.Serialize(file)
	save2 := file.Bytes()

	// XXX: If we start saving the rng seed, with the current (2010-02-14) rng
	// implementation, the seed will end up different on the second save and
	// this will fail.
	if bytes.Compare(save1, save2) != 0 {
		t.Errorf("Second save game made a different save.")
	}
}
