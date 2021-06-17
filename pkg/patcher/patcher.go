package patcher

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
)

type Patch struct {
	id   string
	body []byte
}

type Patcher struct {
	initPatch *Patch
	patches   map[string]*Patch
}

func NewPatcher() *Patcher {
	return &Patcher{
		patches: make(map[string]*Patch),
	}
}

func (p *Patcher) String() string {
	return "patcher"
}

// Rather than creating a reset function, just throw away
// the patcher and create a new one

func (p *Patcher) NewPatch(thePath string) (*Patch, error) {
	return &Patch{}, nil
}

func (p *Patcher) walkDirFunc(thePath string, d fs.DirEntry, err error) error {
	if !d.IsDir() {
		filename := path.Base(thePath)
		if filename == "init_patch.sql" {
			initPatch, err := p.NewPatch(thePath)
			if err != nil {
				return err
			}

			p.initPatch = initPatch
		}
		fmt.Printf("WalkFunc called, path: %q (base: %q), direntry: %v, err: %s\n", thePath, filename, d, err)
	}
	return err
}

func (p *Patcher) Scan(folder string) error {
	filepath.WalkDir(folder, p.walkDirFunc)
	if p.initPatch == nil {
		fmt.Println("******** No init found! ********")
	} else {
		fmt.Println("Here's the patch config:")
		fmt.Println(p)
	}
	return nil
}
