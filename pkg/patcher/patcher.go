/*
Copyright © 2021 Colin Fox <greenenergy@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
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
	initPatch *patch.Patch
	patches   map[string]*patch.Patch
	ordered   []*patch.Patch
	dry       bool
	verbose   bool
	engine    dbe.DBEngine
}

func GetFlagString(name string, flags *pflag.FlagSet) (string, error) {
	flg := flags.Lookup(name)
	if flg != nil {
		if flg.Value != nil {
			return flg.Value.String(), nil
		}
	}
	return "", fmt.Errorf("flag not found")
}

func NewPatcher(flags *pflag.FlagSet) (*Patcher, error) {
	dry := false
	verbose := false
	var enginename string
	var credsName string
	var engine dbe.DBEngine
	var err error

	if flags != nil {
		dry = flags.Lookup("dry").Value.String() == "true"
		enginename = flags.Lookup("engine").Value.String()

		tmp, err := GetFlagString("dbcreds", flags)
		if err == nil {
			credsName = tmp
		}

		verbose = flags.Lookup("verbose").Value.String() == "true"
	}

	switch enginename {
	case "":
		engine = dbe.NewMockDBE(flags)

	case "postgres":
		engine, err = dbe.NewPGDBE(credsName, verbose)
		if err != nil {
			return nil, err
		}

	case "sqlite":
		engine, err = dbe.NewSQLiteDBE(credsName, verbose)
		if err != nil {
			return nil, err
		}
	}

	return &Patcher{
		dry:     dry,
		verbose: verbose,
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
		val := strings.Trim(parts[2], " ")
		switch key {
		case "id":
			newp.Id = val
		case "prereqs":
			newp.Prereqs = strings.Split(val, ",")
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

		if p.verbose {
			fmt.Println("walkdir, checking:", thePath)
		}

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
	// Make sure to apply the init_patch.sql file first
	if ids.Len() == 0 {
		if err = p.engine.Patch(p.initPatch); err != nil {
			return err
		}
		// Skip applying this in the following loop
		ids.Add(p.initPatch.Id)
	}

	numDone := 0
	for _, thepatch := range p.ordered {
		if ids.Contains(thepatch.Id) {
			continue
		}
		if p.verbose || p.dry {
			fmt.Printf("applying: (weight %d) %s\n", thepatch.Weight, thepatch.Filename)
		}

		if !p.dry {
			if err := p.engine.Patch(thepatch); err != nil {
				return err
			}
		}
		numDone += 1
	}
	if numDone > 0 {
		fmt.Printf("Applied %d patches\n", numDone)
	}

	return nil
}
