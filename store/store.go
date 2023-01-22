package store

import (
	"encoding/json"
	"io"
	"os"
)

type Store struct {
	Path string `json:"path"`
	R    uint   `json:"R"`
	G    uint   `json:"G"`
	B    uint   `json:"B"`
}

func Load(path string) (*[]Store, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteVal, _ := io.ReadAll(jsonFile)

	var store []Store
	if err := json.Unmarshal(byteVal, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func Save(path string, stor []Store) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	byteVal, err := json.Marshal(stor)
	if err != nil {
		return err
	}

	_, err = f.Write(byteVal)
	return err
}
