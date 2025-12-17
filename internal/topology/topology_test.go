package topology

import "testing"

func TestSelectOSAndGame(t *testing.T) {
	osCPUs, gameCPUs, lists, err := SelectOSAndGame([]string{"0-3", "4-7"})
	if err != nil {
		t.Fatalf("SelectOSAndGame: %v", err)
	}
	if osCPUs != "0-3" {
		t.Fatalf("unexpected os: %q", osCPUs)
	}
	if gameCPUs != "4-7" {
		t.Fatalf("unexpected game: %q", gameCPUs)
	}
	if len(lists) != 2 {
		t.Fatalf("unexpected lists: %v", lists)
	}
}
