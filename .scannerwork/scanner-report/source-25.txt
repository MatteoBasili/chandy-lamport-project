package utils_test

import (
	"chandy_lamport/src/utils"
	"testing"
)

func TestNewAppMsg(t *testing.T) {
	id := "123"
	body := 42
	from := 1
	to := 2

	msg := utils.NewAppMsg(id, body, from, to)

	if msg.Msg.ID != id {
		t.Errorf("Expected ID %s, but got %s", id, msg.Msg.ID)
	}

	if msg.Msg.Body != body {
		t.Errorf("Expected Body %d, but got %d", body, msg.Msg.Body)
	}

	if msg.IsMarker {
		t.Error("Expected IsMarker to be false, but got true")
	}

	if msg.From != from {
		t.Errorf("Expected From %d, but got %d", from, msg.From)
	}

	if msg.To != to {
		t.Errorf("Expected To %d, but got %d", to, msg.To)
	}
}

func TestNewMarkMsg(t *testing.T) {
	from := 3

	msg := utils.NewMarkMsg(from)

	if msg.Msg.ID != "" {
		t.Error("Expected ID to be empty string, but got non-empty string")
	}

	if msg.Msg.Body != 0 {
		t.Error("Expected Body to be 0, but got non-zero value")
	}

	if !msg.IsMarker {
		t.Error("Expected IsMarker to be true, but got false")
	}

	if msg.From != from {
		t.Errorf("Expected From %d, but got %d", from, msg.From)
	}

	if msg.To != -1 {
		t.Error("Expected To to be -1, but got a different value")
	}
}
