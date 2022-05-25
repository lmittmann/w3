package internal

import (
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEquateErrors(t *testing.T) {
	tests := []struct {
		A, B      interface{}
		WantEqual bool
	}{
		{
			A:         fmt.Errorf("err"),
			B:         fmt.Errorf("err"),
			WantEqual: true,
		},
		{
			A:         fmt.Errorf("err: err2"),
			B:         fmt.Errorf("err: %w", fmt.Errorf("err2")),
			WantEqual: true,
		},
		{
			A:         fmt.Errorf("EOF"),
			B:         io.EOF,
			WantEqual: true,
		},
		{
			A:         fmt.Errorf("err"),
			B:         fmt.Errorf("xxx"),
			WantEqual: false,
		},
		{
			A:         fmt.Errorf("err"),
			B:         nil,
			WantEqual: false,
		},
		{
			A:         nil,
			B:         fmt.Errorf("err"),
			WantEqual: false,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if gotEqual := cmp.Equal(test.A, test.B, EquateErrors()); test.WantEqual != gotEqual {
				t.Fatalf("want %t, got %t", test.WantEqual, gotEqual)
			}
		})
	}
}
