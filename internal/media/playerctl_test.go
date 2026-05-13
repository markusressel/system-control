package media

import (
	"reflect"
	"testing"
)

func TestMatchPlayersExactMatchPreferred(t *testing.T) {
	players := []string{"spotify", "firefox", "spotifyd"}

	matches, err := MatchPlayers(players, "spotify")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"spotify"}
	if !reflect.DeepEqual(matches, expected) {
		t.Fatalf("expected %v, got %v", expected, matches)
	}
}

func TestMatchPlayersRegex(t *testing.T) {
	players := []string{"spotify", "firefox", "spotifyd"}

	matches, err := MatchPlayers(players, "spot.*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"spotify", "spotifyd"}
	if !reflect.DeepEqual(matches, expected) {
		t.Fatalf("expected %v, got %v", expected, matches)
	}
}

func TestMatchPlayersInvalidRegex(t *testing.T) {
	players := []string{"spotify"}

	_, err := MatchPlayers(players, "(")
	if err == nil {
		t.Fatal("expected regex error, got nil")
	}
}
