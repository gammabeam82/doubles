package utils

import (
	. "doubles/types"
	"flag"
	"fmt"
	"strings"
)

func InArray(search string, array []string) bool {
	for _, v := range array {
		if v == search {
			return true
		}
	}
	return false
}

func GetCliOptions() (*Options, error) {
	options := &Options{}

	flag.StringVar(&options.Directory, "dir", "", "Path to directory")
	flag.BoolVar(&options.Delete, "delete", false, "Delete doubles")
	flag.BoolVar(&options.Dump, "dump", false, "Save dump to file")
	skip := flag.String("skip", "", "Comma separated list of subdirectories to skip")
	flag.Parse()
	options.Skip = strings.Split(*skip, ",")

	if len(options.Directory) < 1 {
		fmt.Print("Enter path to directory: ")
		if _, err := fmt.Scan(&options.Directory); err != nil {
			return nil, err
		}
	}

	return options, nil
}
