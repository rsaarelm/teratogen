package teratogen

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWorld(t *testing.T) {
	InitWorld()
	world := GetWorld()
	if world == nil {
		t.Errorf("World not initialized.")
	}

	world.InitLevel(1)

	if world.GetPlayer() == nil {
		t.Errorf("Player not initialized.")
	}

	// Store a string representation of player state for later.
	playerState := fmt.Sprintf("%#v", world.GetPlayer())

	// Try saving the game.
	file := new(bytes.Buffer)
	SaveGame(file)

	// Munge stuff a bit.
	world.InitLevel(2)

	LoadGame(bytes.NewBuffer(file.Bytes()))

	// See that at least some of the state is the same.
	if playerState != fmt.Sprintf("%#v", world.GetPlayer()) {
		t.Errorf("Player state was changed during save-load.")
	}
}
