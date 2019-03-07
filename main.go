package main

import (
	"doubles/config"
	"doubles/doubles"
	. "doubles/types"
	"doubles/utils"
	"fmt"
	"log"
	"time"

	colors "github.com/logrusorgru/aurora"
)

const pathToConfig string = "./config/config.json"

var conf *Config

func init() {
	conf = &Config{}
	if err := config.LoadConfig(pathToConfig, conf); err != nil {
		log.Fatal(colors.Red(err))
	}
}

func main() {
	options, err := utils.GetCliOptions()
	if err != nil {
		log.Fatal(colors.Red(err))
	}

	start := time.Now()

	doubles.Run(options, conf)

	duration := time.Since(start)
	fmt.Printf("\nDone in: %s\n", colors.Green(duration))
}
