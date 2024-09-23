//go:generate go run gen.go

package fourbyte

import "github.com/lmittmann/w3"

func Function(sig [4]byte) *w3.Func {
	return functions[sig]
}

func Event(topic0 [32]byte) *w3.Event {
	return events[topic0]
}
