// Package data comment
package data

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"strings"
)

// testDataDir a directory container test data.
type testDataDir struct {
	prefix string
	fs     embed.FS
}

//go:embed tx/bin/*
var txBinData embed.FS

// TxBinData data for binary txs.
var TxBinData = testDataDir{
	prefix: "tx/bin",
	fs:     txBinData,
}

// Open a file.
func (d *testDataDir) Open(file string) (fs.File, error) {
	return d.fs.Open(path.Join(d.prefix, file))
}

// Load the data of a file.
func (d *testDataDir) Load(file string) ([]byte, error) {
	return d.fs.ReadFile(path.Join(d.prefix, file))
}

// GetTestHex is a convenience function for reading local .hex transaction files, returnins a string value
func GetTestHex(fileName string) string {
	fileData, err := os.ReadFile(fileName) //nolint:gosec // only used in testing
	if err != nil {
		return ""
	}

	return strings.Trim(string(fileData), "\n")
}
