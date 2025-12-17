package topology

import (
	"reflect"
	"testing"
)

func TestParseAndFormatCPUList(t *testing.T) {
	parsed, err := ParseCPUList("0-2,4, 6-7,7")
	if err != nil {
		t.Fatalf("ParseCPUList: %v", err)
	}
	if want := []int{0, 1, 2, 4, 6, 7}; !reflect.DeepEqual(parsed, want) {
		t.Fatalf("unexpected parse: got=%v want=%v", parsed, want)
	}
	if got := FormatCPUList(parsed); got != "0-2,4,6-7" {
		t.Fatalf("unexpected format: %q", got)
	}
}

func TestParseCPUList_Invalid(t *testing.T) {
	if _, err := ParseCPUList("3-1"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := ParseCPUList("x"); err == nil {
		t.Fatalf("expected error")
	}
}
