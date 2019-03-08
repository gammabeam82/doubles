package config

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
)

type Loader func(c *Config) error

type Config struct {
	ImageTypes []string `json:"image_types" xml:"image-type"`
}

func (c *Config) Load(loader Loader) error {
	return loader(c)
}

func NewConfig() *Config {
	return &Config{}
}

func NewJsonLoader(filename string) Loader {
	return func(conf *Config) error {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, conf); err != nil {
			return err
		}
		return nil
	}
}

func NewXMLLoader(filename string) Loader {
	return func(conf *Config) error {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err = xml.Unmarshal(data, conf); err != nil {
			return err
		}
		return nil
	}
}
