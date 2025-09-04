package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
	"gopkg.in/yaml.v3"
)

type CategoryConfig struct {
	Sheet 			string				`yaml:"sheet"`
	SKUHeader 		string				`yaml:"sku_header"`
	Headers 		[]string			`yaml:"headers"`
	FieldMapping 	map[string]string	`yaml:"field_mapping"`
}

var sheets map[string]string

func init() {
    b, err := os.ReadFile("configs/sheets.yaml")
    if err != nil {
        // нельзя logger.Log здесь — он еще nil
        fmt.Fprintf(os.Stderr, "failed to read .yaml file: %v\n", err)
        sheets = map[string]string{}
        return
    }
    if err := yaml.Unmarshal(b, &sheets); err != nil {
        fmt.Fprintf(os.Stderr, "failed to unmarshal yaml: %v\n", err)
        sheets = map[string]string{}
        return
    }
}

func LoadBySlug(slug string) (CategoryConfig, error) {
	path, ok := sheets[slug]
	if !ok {
		txt := fmt.Sprintf("unknown slug: %s", slug)
		err := errors.New(txt)
		logger.Log.Errorf("failed to get path: %v", err)
		return CategoryConfig{}, err
	}

	return Load(path)
}

func Load(path string) (CategoryConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Errorf("failed to read file path: %v", err)
		return CategoryConfig{}, err
	}

	var c CategoryConfig
	if err := yaml.Unmarshal(b, &c); err != nil {
		return CategoryConfig{}, err
	}

	return c, nil
}

