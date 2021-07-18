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

	"github.com/greenenergy/dbp/pkg/dbe"
	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/spf13/pflag"
)

type Patcher struct {
	initPatch      *patch.Patch
	patches        map[string]*patch.Patch
	alreadyPatched map[string]bool
	ordered        []*patch.Patch
	dry            bool
	engine         dbe.DBEngine
}

func NewPatcher(flags *pflag.FlagSet) (*Patcher, error) {
	dry := false
	var enginename string
	var engine dbe.DBEngine
	var err error

	if flags != nil {
		dry = flags.Lookup("dry").Value.String() == "true"
		enginename = flags.Lookup("engine").Value.String()
	}

	credsName := flags.Lookup("dbcreds").Value.String()

	switch enginename {
	case "":
		engine = dbe.NewMockDBE(flags)

	case "postgres":
		engine, err = dbe.NewPGDBE(credsName)
		if err != nil {
			return nil, err
		}

	case "sqlite":
		engine = dbe.NewSQLiteDBE(credsName)
	}

	return &Patcher{
		dry:     dry,
		patches: make(map[string]*patch.Patch),
		engine:  engine,
	}, nil
}

func (p *Patcher) String() string {
	dummy, _ := json.MarshalIndent(p.ordered, "", "    ")
	return string(dummy)
}

// Dry - Manually set the dryrun flag, mainly for testing
func (p *Patcher) Dry(dry bool) {
	p.dry = dry
}

// Rather than creating a reset function, just throw away
// the patcher and create a new one

func (p *Patcher) NewPatch(thePath string) (*patch.Patch, error) {
	// Need to add the scanning & interpretation code here.

	file, err := os.Open(thePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	newp := patch.Patch{
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

	if newp.Id == "" {
		// The ID field is an absolute must. Without it, there is
		// no linking.
		return nil, fmt.Errorf("file %q missing ID field", thePath)
	}

	data, err := ioutil.ReadFile(thePath)
	if err != nil {
		return nil, err
	}
	newp.Body = data

	if other, ok := p.patches[newp.Id]; ok {
		return nil, fmt.Errorf("duplicate id: %s, found in File %q and %q", newp.Id, other.Filename, newp.Filename)
	}
	p.patches[newp.Id] = &newp
	return &newp, nil
}

// bumpWeight - the recursive weighting function that implements the prerequisite linking
func (p *Patcher) bumpWeight(thepatch *patch.Patch, detectionMap map[string]*patch.Patch) error {

	if _, ok := detectionMap[thepatch.Id]; ok {
		var filenames []string
		for key := range detectionMap {
			filenames = append(filenames, p.patches[key].Filename)
		}
		return fmt.Errorf("loop detected:\n%s", strings.Join(filenames, "\n"))
	}

	thepatch.Weight += 1
	detectionMap[thepatch.Id] = thepatch

	for _, patchkey := range thepatch.Prereqs {

		err := p.bumpWeight(p.patches[patchkey], detectionMap)
		if err != nil {
			return err
		}
	}
	delete(detectionMap, thepatch.Id)
	return nil
}

func (p *Patcher) Resolve() error {
	var patches []*patch.Patch

	for _, thispatch := range p.patches {
		patches = append(patches, thispatch)
		detectionMap := make(map[string]*patch.Patch)
		detectionMap[thispatch.Id] = thispatch

		for _, pre := range thispatch.Prereqs {
			if other, ok := p.patches[pre]; !ok {
				// TODO: It's possible to reference a patch that has already been applied. So if the patch is not found
				// in the list of files, we should check the database to see if that ID already exists in the 'applied_patches' table,
				// and if it does, then let this trhough. Possibly create a mock entry so the rest of the code works as expected,
				// and we can flag this as "applied patch found, though no file exists for it" for forensic examination.
				return fmt.Errorf("bad ID reference. File %q refers to id %q which doesn't exist", thispatch.Filename, pre)
			} else {
				err := p.bumpWeight(other, detectionMap)
				if err != nil {
					return err
				}
			}
		}
	}

	sort.Sort(patch.ByWeight(patches))
	p.ordered = patches
	return nil
}

func (p *Patcher) walkDirFunc(thePath string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !strings.HasSuffix(strings.ToLower(thePath), ".sql") {
		// skip anything that doesn't end with 'sql'
		return nil
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
	ids, err := p.engine.GetInstalledIDs()
	if err != nil {
		return err
	}

	for _, thepatch := range p.ordered {
		if ids.Contains(thepatch.Id) {
			continue
		}
		if p.dry {
			fmt.Printf("would apply (weight %d): ", thepatch.Weight)
		}

		if err := p.engine.Patch(thepatch); err != nil {
			return err
		}
	}
	return nil
}
