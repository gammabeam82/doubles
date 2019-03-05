package main

import (
	"crypto/md5"
	. "doubles/types"
	"doubles/utils"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	colors "github.com/logrusorgru/aurora"
	"github.com/schollz/progressbar"
)

var (
	imageTypes = []string{
		"image/jpeg",
		"image/png",
		"image/gif",
	}
	wg     sync.WaitGroup
	images = NewImageCollection()
)

func isImage(file *os.File) (bool, error) {
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return false, err
	}
	mimeType := http.DetectContentType(buffer)
	return utils.InArray(mimeType, imageTypes), nil
}

func isPathValid(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func calculateHash(files <-chan string, results chan<- struct{}) {
	for filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(colors.Red(err))
		}

		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			log.Fatal(colors.Red(err))
		}

		if err := file.Close(); err != nil {
			log.Println(err)
		}

		images.AddHash(hash.Sum(nil), filename)
		results <- struct{}{}
	}
}

func scan(dir string, skip []string) {
	defer wg.Done()

	visit := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if utils.InArray(path.Base(currentPath), skip) {
			return filepath.SkipDir
		}

		if info.IsDir() && currentPath != dir {
			wg.Add(1)
			go scan(currentPath, skip)
			return filepath.SkipDir
		}

		if !info.IsDir() && info.Mode().IsRegular() {
			file, err := os.Open(currentPath)
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil {
					log.Println(err)
				}
			}()
			isImg, err := isImage(file)
			if err != nil {
				return err
			}
			if isImg {
				images.AddFile(currentPath)
			}
		}
		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		log.Fatal(colors.Red(err))
	}
}

func deleteAllExceptFirst(doubles *map[string]Doubles) (int, error) {
	num := 0
	for _, list := range *doubles {
		for _, filename := range list[1:] {
			if err := os.Remove(filename); err != nil {
				return 0, err
			}
			num++
		}
	}
	return num, nil
}

func getCliOptions() *Options {
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
			log.Fatal(colors.Red(err))
		}
	}

	return options
}

func run() {
	options := getCliOptions()

	if !isPathValid(options.Directory) {
		log.Fatal(colors.Red("Invalid path"))
	}

	fmt.Println("Scanning directory... ")

	wg.Add(1)
	scan(options.Directory, options.Skip)
	wg.Wait()

	length := images.Length()
	fmt.Printf("Images found: %d\n", colors.Green(length))

	if length == 0 {
		return
	}

	jobs := make(chan string, length)
	results := make(chan struct{}, length)

	defer func() {
		close(jobs)
		close(results)
	}()

	fmt.Println("Calculating hashes... ")
	bar := progressbar.New(length)

	for w := 1; w <= 50; w++ {
		go calculateHash(jobs, results)
	}

	for _, filename := range images.Files() {
		jobs <- filename
	}

	for i := 1; i <= length; i++ {
		<-results
		if err := bar.Add(1); err != nil {
			log.Println(err)
		}
	}

	doubles := images.FindDoubles()
	fmt.Printf("\n\nDoubles found: %d\n", colors.Green(len(doubles)))

	if options.Dump {
		data, _ := json.MarshalIndent(doubles, "", "\t")
		if err := ioutil.WriteFile("dump.json", data, 0644); err != nil {
			log.Println(err)
		}
	}

	for _, list := range doubles {
		fmt.Println(list)
	}

	if options.Delete {
		num, err := deleteAllExceptFirst(&doubles)
		if err != nil {
			log.Fatal(colors.Red(err))
		}
		fmt.Printf("\n\nDeleted %d file(s)\n", colors.Bold(colors.Red(num)))
	}
}

func main() {
	start := time.Now()

	run()

	end := time.Since(start)
	fmt.Printf("\nDone in: %s\n", colors.Green(end))
}
