package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
)

type Loader func(c *Config) error

type Config struct {
	ImageTypes []string `json:"image_types" xml:"image-type"`
	DumpFile   string   `json:"dump_file" xml:"dump-file"`
}

func (c *Config) Load(loader Loader) error {
	return loader(c)
}

func (c *Config) validate() error {
	if len(c.ImageTypes) > 0 && len(c.DumpFile) > 0 {
		return nil
	}
	return errors.New("Invalid config")
}

func NewConfig() *Config {
	return &Config{}
}

func NewJsonLoader(filename string) Loader {
	return func(c *Config) error {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, c); err != nil {
			return err
		}
		return c.validate()
	}
}

func NewXMLLoader(filename string) Loader {
	return func(c *Config) error {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err = xml.Unmarshal(data, c); err != nil {
			return err
		}
		return c.validate()
	}
}
