package patcher

import (
	"bufio"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Patch struct {
	Id          string
	Patch       string
	Author      string
	Description string
	Prereqs     []string
	body        []byte
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
	// Need to add the scanning & interpretation code here.

	file, err := os.Open(thePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	newp := Patch{
		Id: uuid.New().String(),
	}

	keyfilter := regexp.MustCompile(`^-- ([\w\d]+): ([\w\d\s.,@-]+)$`)
	for scanner.Scan() {
		s := scanner.Text()
		parts := keyfilter.FindStringSubmatch(s)
		// parts[0] == whole line
		// parts[1] == keyword
		// parts[2] == value
		if len(parts) < 2 {
			continue
		}

		key := parts[1]
		val := parts[2]
		switch key {
		case "PATCH":
			newp.Patch = val
		case "id":
			newp.Id = val
		case "author":
			newp.Author = val
		case "prereqs":
			newp.Prereqs = strings.Split(val, ",")
		case "description":
			newp.Description = val
		}
	}

	data, err := ioutil.ReadFile(thePath)
	if err != nil {
		return nil, err
	}
	newp.body = data

	//dummy, _ := json.MarshalIndent(newp, "", "    ")
	//fmt.Println(string(dummy))

	p.patches[newp.Id] = &newp

	return &newp, nil
}

func (p *Patcher) walkDirFunc(thePath string, d fs.DirEntry, err error) error {
	if !d.IsDir() {
		filename := path.Base(thePath)
		initPatch, err := p.NewPatch(thePath)

		if filename == "init_patch.sql" {
			if err != nil {
				return err
			}

			p.initPatch = initPatch
		}
		//fmt.Printf("WalkFunc called, path: %q (base: %q), direntry: %v, err: %q\n", thePath, filename, d, err)
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
