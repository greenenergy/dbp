package patcher

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

type Patch struct {
	Id          string
	Patch       string
	Author      string
	Description string
	Tags        []string
	Prereqs     []string
	body        []byte
	Weight      int
	Filename    string
}

type ByWeight []*Patch

func (by ByWeight) Len() int {
	return len(by)
}

func (by ByWeight) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

func (by ByWeight) Less(i, j int) bool {
	return by[i].Weight > by[j].Weight
}

type Patcher struct {
	initPatch *Patch
	patches   map[string]*Patch
	ordered   []*Patch
	dry       bool
}

func NewPatcher(flags *pflag.FlagSet) *Patcher {

	return &Patcher{
		dry:     flags.Lookup("dry").Value.String() == "true",
		patches: make(map[string]*Patch),
	}
}

func (p *Patcher) String() string {
	dummy, _ := json.MarshalIndent(p.ordered, "", "    ")
	return string(dummy)
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
		//Id:       uuid.New().String(),
		Filename: thePath,
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
		case "tags":
			newp.Tags = strings.Split(val, ",")
		case "description":
			newp.Description = val
		}
	}
	//fmt.Printf("file: %q, id: %q\n", thePath, newp.Id)
	if newp.Id == "" {
		// The ID field is an absolute must. Without it, there is
		// no linking.
		return nil, fmt.Errorf("file %q missing ID field", thePath)
	}

	data, err := ioutil.ReadFile(thePath)
	if err != nil {
		return nil, err
	}
	newp.body = data

	if other, ok := p.patches[newp.Id]; ok {
		return nil, fmt.Errorf("duplicate id: %s, found in File %q and %q", newp.Id, other.Filename, newp.Filename)
	}
	p.patches[newp.Id] = &newp
	return &newp, nil
}

// bumpWeight - the recursive weighting function that implements the prerequisite linking
func (p *Patcher) bumpWeight(depth int, patch *Patch, detectionMap map[string]*Patch) error {
	_ = depth
	/*
		for x := 0; x < depth; x++ {
			fmt.Printf(".")
		}
		fmt.Println(patch.Filename)
	*/

	if _, ok := detectionMap[patch.Id]; ok {
		var filenames []string
		for key := range detectionMap {
			filenames = append(filenames, p.patches[key].Filename)
		}
		return fmt.Errorf("loop detected:\n%s", strings.Join(filenames, "\n"))
	}

	patch.Weight += 1
	detectionMap[patch.Id] = patch

	for _, patchkey := range patch.Prereqs {

		err := p.bumpWeight(depth+1, p.patches[patchkey], detectionMap)
		if err != nil {
			return err
		}
	}
	delete(detectionMap, patch.Id)
	return nil
}

func (p *Patcher) Resolve() error {
	var patches []*Patch

	for _, patch := range p.patches {
		patches = append(patches, patch)
		detectionMap := make(map[string]*Patch)
		detectionMap[patch.Id] = patch

		for _, pre := range patch.Prereqs {
			if other, ok := p.patches[pre]; !ok {
				// TODO: It's possible to reference a patch that has already been applied. So if the patch is not found
				// in the list of files, we should check the database to see if that ID already exists in the 'applied_patches' table,
				// and if it does, then let this trhough. Possibly create a mock entry so the rest of the code works as expected,
				// and we can flag this as "applied patch found, though no file exists for it" for forensic examination.
				return fmt.Errorf("bad ID reference. File %q refers to id %q which doesn't exist", patch.Filename, pre)
			} else {
				err := p.bumpWeight(0, other, detectionMap)
				if err != nil {
					return err
				}
			}
		}
	}

	sort.Sort(ByWeight(patches))
	p.ordered = patches
	return nil
}

func (p *Patcher) walkDirFunc(thePath string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		filename := path.Base(thePath)
		initPatch, err := p.NewPatch(thePath)
		if err != nil {
			return err
		}

		if filename == "init_patch.sql" {
			if err != nil {
				return err
			}

			p.initPatch = initPatch
		}
	}
	return err
}

func (p *Patcher) Scan(folder string) error {
	err := filepath.WalkDir(folder, p.walkDirFunc)
	if err != nil {
		return err
	}
	if p.initPatch == nil {
		return errors.New("no init found. There must be at least one file in the patch tree named 'init_patch.sql'")
	} else {
		return p.Resolve()
	}
}

func (p *Patcher) Process() error {
	for _, patch := range p.ordered {
		if p.dry {
			fmt.Printf("would apply (weight %d): ", patch.Weight)
		}
		fmt.Println(patch.Filename)
	}
	return nil
}
