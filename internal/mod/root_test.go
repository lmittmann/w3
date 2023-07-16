package mod_test

import (
	"strings"
	"testing"

	"github.com/lmittmann/w3/internal/mod"
)

func TestModRoot(t *testing.T) {
	if !strings.HasSuffix(mod.Root, "w3") {
		t.Fatalf("Unexpected module root: %q", mod.Root)
	}
}
