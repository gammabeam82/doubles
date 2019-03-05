package main

import (
	"doubles/doubles"
	"doubles/utils"
	"fmt"
	"log"
	"time"

	colors "github.com/logrusorgru/aurora"
)

func main() {
	options, err := utils.GetCliOptions()
	if err != nil {
		log.Fatal(colors.Red(err))
	}

	start := time.Now()

	doubles.Run(options)

	duration := time.Since(start)
	fmt.Printf("\nDone in: %s\n", colors.Green(duration))
}
