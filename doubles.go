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
	"sync"
	"time"

	colors "github.com/logrusorgru/aurora"
	"github.com/schollz/progressbar"
)

type Doubles []string

func (d Doubles) String() string {
	var res string
	for k, v := range d {
		res += v
		if k != len(d)-1 {
			res += fmt.Sprintf("%s", colors.Red("|"))
		}
	}
	return res
}

type ImageCollection struct {
	mux    sync.Mutex
	files  []string
	hashes map[string][]string
}

func (i *ImageCollection) Length() int {
	return len(i.files)
}

func (i *ImageCollection) AddFile(filename string) {
	i.mux.Lock()
	defer i.mux.Unlock()
	i.files = append(i.files, filename)
}

func (i *ImageCollection) AddHash(hash []byte, filename string) {
	i.mux.Lock()
	defer i.mux.Unlock()
	filehash := fmt.Sprintf("%x", hash)
	i.hashes[filehash] = append(i.hashes[filehash], filename)
}

func (i *ImageCollection) FindDoubles() map[string]Doubles {
	doubles := make(map[string]Doubles)
	for k, v := range i.hashes {
		if len(v) > 1 {
			doubles[k] = v
		}
	}
	return doubles
}

func NewImageCollection() *ImageCollection {
	return &ImageCollection{
		hashes: make(map[string][]string),
	}
}

var (
	imageTypes = [...]string{
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

		file.Close()
		images.AddHash(hash.Sum(nil), filename)
		results <- struct{}{}
	}
}

func getFilesList(dir string) error {
	defer wg.Done()

	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			wg.Add(1)
			go getFilesList(path)
			return filepath.SkipDir
		}

		if !info.IsDir() && info.Mode().IsRegular() {
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
				images.AddFile(path)
			}
		}
		return nil
	}

	filepath.Walk(dir, visit)
	return nil
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

func run() {
	var dir string

	fdir := flag.String("dir", "", "Path to directory")
	del := flag.Bool("delete", false, "Delete doubles")
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
	wg.Add(1)
	err := getFilesList(dir)
	wg.Wait()
	if err != nil {
		log.Fatal(colors.Red(err))
	}

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

	for _, filename := range images.files {
		jobs <- filename
	}

	for i := 1; i <= length; i++ {
		<-results
		bar.Add(1)
	}

	doubles := images.FindDoubles()
	fmt.Printf("\n\nDoubles found: %d\n", colors.Green(len(doubles)))

	for _, list := range doubles {
		fmt.Println(list, "\n")
	}

	if *del == true {
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
