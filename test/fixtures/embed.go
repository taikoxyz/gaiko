package fixtures

import (
	"embed"
	"io/fs"
	"path/filepath"
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

func GetBatchInputs() ([][]byte, error) {
	var inputs [][]byte
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
			inputs = append(inputs, file)
			return nil
		},
	)
}
