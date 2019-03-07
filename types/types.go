package types

import (
	"fmt"
	"sync"

	colors "github.com/logrusorgru/aurora"
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

type Options struct {
	Directory string
	Delete    bool
	Dump      bool
	Skip      []string
}

type Config struct {
	ImageTypes []string `json:"image_types"`
}

type ImageCollection struct {
	mux    sync.Mutex
	files  []string
	hashes map[string][]string
}

func (i *ImageCollection) Length() int {
	return len(i.files)
}

func (i *ImageCollection) Files() []string {
	return i.files
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

func (i *ImageCollection) FindDoubles() (int, map[string]Doubles) {
	num := 0
	doubles := make(map[string]Doubles)
	for k, v := range i.hashes {
		if len(v) > 1 {
			doubles[k] = v
			num += len(v)
		}
	}
	return num, doubles
}

func NewImageCollection() *ImageCollection {
	return &ImageCollection{
		hashes: make(map[string][]string),
	}
}
