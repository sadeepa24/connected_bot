package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/sadeepa24/connected_bot/botapi"
	"gopkg.in/yaml.v3"
)

func readTmpl(path string) (map[string]map[string]botapi.MgItem, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var templates map[string]map[string]botapi.MgItem

	switch {
	case strings.Contains(path, ".yaml"):
		return templates, yaml.Unmarshal(file, &templates)
	case strings.Contains(path, ".json"):
		return templates, json.Unmarshal(file, &templates)
	}

	return nil, errors.New("invalid template file type")

}
