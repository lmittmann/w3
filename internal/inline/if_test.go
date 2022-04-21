package inline

import "testing"

func TestIf(t *testing.T) {
	if If(true, 1, 0) != 1 {
		t.Error("If(true, 1, 0) != 1")
	}
	if If(false, 1, 0) != 0 {
		t.Error("If(false, 1, 0) != 0")
	}
}
