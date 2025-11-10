// Package fixtures provides embedded test fixtures for various input and output pairs used in testing.
package fixtures

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed single/*
var singleInputs embed.FS

//go:embed shasta/*
var shastaInputs embed.FS

//go:embed batch/*
var batchInputs embed.FS

type Pair struct {
	Input  []byte
	Output []byte
}

const (
	jsonExtension  = ".json"
	fileTypeInput  = "input"
	fileTypeOutput = "output"
	inputPrefix    = fileTypeInput + "-"
	outputPrefix   = fileTypeOutput + "-"
)

func GetSingleInputs() (map[uint64]*Pair, error) {
	return getInputs(singleInputs, "single")
}

func GetBatchInputs() (map[uint64]*Pair, error) {
	return getInputs(batchInputs, "batch")
}

func GetShastaInputs() (map[uint64]*Pair, error) {
	return getInputs(shastaInputs, "shasta")
}

func getInputs(fsys fs.FS, root string) (map[uint64]*Pair, error) {
	inputs := make(map[uint64]*Pair)
	return inputs, fs.WalkDir(
		fsys,
		root,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != jsonExtension {
				return nil
			}
			ty, id, ok := parseFileName(path)
			if !ok {
				return nil
			}
			file, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			pair := inputs[id]
			if pair == nil {
				pair = &Pair{}
				inputs[id] = pair
			}
			switch ty {
			case fileTypeInput:
				pair.Input = file
			case fileTypeOutput:
				pair.Output = file
			}
			return nil
		},
	)
}

func parseFileName(p string) (string, uint64, bool) {
	name := filepath.Base(p)

	switch {
	case strings.HasPrefix(name, inputPrefix):
		id, err := strconv.ParseUint(strings.TrimSuffix(name[len(inputPrefix):], jsonExtension), 10, 64)
		if err != nil {
			return "", 0, false
		}
		return fileTypeInput, id, true
	case strings.HasPrefix(name, outputPrefix):
		id, err := strconv.ParseUint(strings.TrimSuffix(name[len(outputPrefix):], jsonExtension), 10, 64)
		if err != nil {
			return "", 0, false
		}
		return fileTypeOutput, id, true
	default:
		return "", 0, false
	}
}
