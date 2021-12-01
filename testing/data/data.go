package data

import (
	"embed"
	"io/fs"
	"path"
)

// TestDataDir a directory container test data.
type TestDataDir struct {
	prefix string
	fs     embed.FS
}

//go:embed tx/bin/*
var txBinData embed.FS

// TxBinData data for binary txs.
var TxBinData = TestDataDir{
	prefix: "tx/bin",
	fs:     txBinData,
}

// Open a file.
func (d *TestDataDir) Open(file string) (fs.File, error) {
	return d.fs.Open(path.Join(d.prefix, file))
}

// Load the data of a file.
func (d *TestDataDir) Load(file string) ([]byte, error) {
	return d.fs.ReadFile(path.Join(d.prefix, file))
}
