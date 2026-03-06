package data

import (
	"bytes"
	_ "embed"
)

//go:embed simulation.json
var SimulationJSON []byte

// utf8BOM is the 3-byte UTF-8 byte order mark that some editors prepend to files.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func GetSimulationData() []byte {
	// Strip UTF-8 BOM if present — Go's JSON decoder does not tolerate it.
	return bytes.TrimPrefix(SimulationJSON, utf8BOM)
}
