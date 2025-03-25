package fixtures

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed single/*
var singleInputs embed.FS

//go:embed batch/*
var batchInputs embed.FS

func GetSingleInputs() ([][]byte, error) {
	var inputs [][]byte
	return inputs, fs.WalkDir(
		singleInputs,
		"single",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".json" {
				return nil
			}
			file, err := fs.ReadFile(singleInputs, path)
			if err != nil {
				return err
			}
			inputs = append(inputs, file)
			return nil
		},
	)
}

type Pair struct {
	Input  []byte
	Output []byte
}

func GetBatchInputs() (map[uint64]*Pair, error) {
	inputs := map[uint64]*Pair{}
	return inputs, fs.WalkDir(
		batchInputs,
		"batch",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".json" {
				return nil
			}
			file, err := fs.ReadFile(batchInputs, path)
			if err != nil {
				return err
			}
			ty, id := parseFileName(path)
			_, ok := inputs[id]
			if !ok {
				inputs[id] = &Pair{}
			}
			switch ty {
			case "input":
				inputs[id].Input = file
			case "output":
				inputs[id].Output = file
			default:
				return fmt.Errorf("invalid file name: %s", path)
			}
			return nil
		},
	)
}

func parseFileName(p string) (string, uint64) {
	p = filepath.Base(p)
	const inputPrefix = "batch-input-"
	const outputPrefix = "batch-output-"
	if strings.HasPrefix(p, inputPrefix) {
		p = strings.TrimPrefix(p, inputPrefix)
		p = strings.TrimSuffix(p, ".json")
		id, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return "", 0
		}
		return "input", id
	}
	if strings.HasPrefix(p, outputPrefix) {
		p = strings.TrimPrefix(p, outputPrefix)
		p = strings.TrimSuffix(p, ".json")
		id, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return "", 0
		}
		return "output", id
	}
	return "", 0
}
