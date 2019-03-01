package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	colors "github.com/logrusorgru/aurora"
	"github.com/schollz/progressbar"
)

var imageTypes = [...]string{
	"image/jpeg",
	"image/png",
}

type File struct {
	name string
	hash []byte
	size int64
}

func (f File) Hash() string {
	return fmt.Sprintf("%x", f.hash)
}

func (f File) Name() string {
	return f.name
}

func (f File) String() string {
	return fmt.Sprintf(
		"\nSize: %-4dkB\tName: %s",
		colors.Cyan(f.size/1024),
		colors.Green(f.name),
	)
}

func findDoubles(files *map[string][]File) map[string][]File {
	doubles := make(map[string][]File)
	for key, value := range *files {
		if len(value) > 1 {
			doubles[key] = value
		}
	}
	return doubles
}

func isImage(file *os.File) (bool, error) {
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return false, err
	}
	mimeType := http.DetectContentType(buffer)
	for _, t := range imageTypes {
		if t == mimeType {
			return true, nil
		}
	}
	return false, nil
}

func isPathValid(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func calculateHash(files <-chan string, results chan<- File) {
	for filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(colors.Red(err))
		}

		stat, err := file.Stat()
		if err != nil {
			log.Fatal(colors.Red(err))
		}

		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			log.Fatal(colors.Red(err))
		}

		file.Close()
		results <- File{filename, hash.Sum(nil), stat.Size()}
	}
}

func deleteAllExceptFirst(doubles *map[string][]File) (int, error) {
	num := 0
	for _, list := range *doubles {
		for _, file := range list[1:] {
			if err := os.Remove(file.Name()); err != nil {
				return 0, err
			}
			num++
		}
	}
	return num, nil
}

func getFilesList(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		isImg, err := isImage(file)
		if err != nil {
			return err
		}
		if isImg {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func run() {
	var dir string

	fdir := flag.String("dir", "", "Path to directory")
	delete := flag.Bool("delete", false, "Delete doubles")
	flag.Parse()

	if len(*fdir) > 1 {
		dir = *fdir
	} else {
		fmt.Print("Enter path to directory: ")
		fmt.Scan(&dir)
	}

	if !isPathValid(dir) {
		log.Fatal(colors.Red("Invalid path"))
	}

	fmt.Println("Scanning directory... ")
	files, err := getFilesList(dir)
	if err != nil {
		log.Fatal(colors.Red(err))
	}

	length := len(files)
	fmt.Printf("Images found: %d\n", colors.Green(length))
	jobs := make(chan string, length)
	results := make(chan File, length)
	rs := make(map[string][]File)

	defer func() {
		close(jobs)
		close(results)
	}()

	fmt.Println("Calculating hashes... ")
	bar := progressbar.New(length)

	for w := 1; w <= 50; w++ {
		go calculateHash(jobs, results)
	}

	for _, file := range files {
		jobs <- file
	}

	for i := 1; i <= length; i++ {
		file := <-results
		bar.Add(1)
		hash := file.Hash()
		rs[hash] = append(rs[hash], file)
	}

	doubles := findDoubles(&rs)
	for _, list := range doubles {
		fmt.Println(list, "\n")
	}

	if *delete == true {
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
