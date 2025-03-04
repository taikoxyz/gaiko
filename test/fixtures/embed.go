package fixtures

import (
	"embed"
	"encoding/json"
	"io/fs"
	"path/filepath"

	"github.com/taikoxyz/gaiko/internal/witness"
)

//go:embed single/*
var singleInputs embed.FS

//go:embed batch/*
var batchInputs embed.FS

func GetSingleInputs() ([]*witness.GuestInput, error) {
	var inputs []*witness.GuestInput
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
			var input witness.GuestInput
			if err := json.Unmarshal(file, &input); err != nil {
				return err
			}
			inputs = append(inputs, &input)
			return nil
		},
	)
}

func GetBatchInputs() ([]*witness.BatchGuestInput, error) {
	var inputs []*witness.BatchGuestInput
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
			var input witness.BatchGuestInput
			if err := json.Unmarshal(file, &input); err != nil {
				return err
			}
			inputs = append(inputs, &input)
			return nil
		},
	)
}
