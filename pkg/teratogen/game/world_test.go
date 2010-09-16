package game

import (
	"bytes"
	"hyades/entity"
	"testing"
)

func TestWorld(t *testing.T) {
	NewContext().InitGame()

	if PlayerId() == entity.NilId {
		t.Errorf("Player not initialized.")
	}

	// Store a string representation of player state for later.
	playerPos := GetPos(PlayerId())

	// Try saving the game.
	file := new(bytes.Buffer)
	GetContext().Serialize(file)

	// Munge stuff a bit.
	NextLevel()

	// Load game.
	GetContext().Deserialize(bytes.NewBuffer(file.Bytes()))

	// See that at least some of the state is the same.
	if !playerPos.Equals(GetPos(PlayerId())) {
		t.Errorf("Player state was changed during save-load.")
	}

	// Save the game again, see if the serializations are same.
	file2 := new(bytes.Buffer)
	GetContext().Serialize(file2)

	if !bytes.Equal(file.Bytes(), file2.Bytes()) {
		t.Errorf("Different save data after loading and re-saving.")
	}
}
