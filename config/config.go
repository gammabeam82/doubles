package config

import (
	. "doubles/types"
	"encoding/json"
	"io/ioutil"
)

func LoadConfig(filename string, config *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, config); err != nil {
		return err
	}
	return nil
}
