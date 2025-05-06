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

//go:embed batch/*
var batchInputs embed.FS

type Pair struct {
	Input  []byte
	Output []byte
}

func GetSingleInputs() (map[uint64]*Pair, error) {
	return getInputs(singleInputs, "single")
}

func GetBatchInputs() (map[uint64]*Pair, error) {
	return getInputs(batchInputs, "batch")
}

func getInputs(fsys fs.FS, root string) (map[uint64]*Pair, error) {
	inputs := map[uint64]*Pair{}
	return inputs, fs.WalkDir(
		fsys,
		root,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".json" {
				return nil
			}
			file, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			ty, id := parseFileName(path)
			switch ty {
			case "input":
				if inputs[id] == nil {
					inputs[id] = &Pair{}
				}
				inputs[id].Input = file
			case "output":
				if inputs[id] == nil {
					inputs[id] = &Pair{}
				}
				inputs[id].Output = file
			default:
				return nil
			}
			return nil
		},
	)
}

func parseFileName(p string) (string, uint64) {
	p = filepath.Base(p)
	const inputPrefix = "input-"
	const outputPrefix = "output-"
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
