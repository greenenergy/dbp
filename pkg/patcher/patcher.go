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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/greenenergy/dbp/pkg/dbe"
	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/spf13/pflag"
)

type Patcher struct {
	initPatch    *patch.Patch
	patches      map[string]*patch.Patch
	ordered      []*patch.Patch
	dry          bool
	verbose      bool
	engine       dbe.DBEngine
	installedIDs *set.Set
	folder       string
	ignore       string
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

func NewPatcher(dry, verbose bool, engine dbe.DBEngine, folder, ignore string) (*Patcher, error) {
	return &Patcher{
		dry:     dry,
		verbose: verbose,
		patches: make(map[string]*patch.Patch),
		engine:  engine,
		folder:  folder,
		ignore:  ignore,
	}, nil
}

func (p *Patcher) Reset() {
	p.ordered = []*patch.Patch{}
	p.patches = make(map[string]*patch.Patch)
	p.initPatch = nil
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

var headerLine = regexp.MustCompile(`^-- ([\w\d]+): ([\w\d\s.,@-]+)$`)

// parseHeader fills in the patch's fields from the comment lines at the top of
// the file. It stops at the first line of SQL, and the first occurrence of a
// key wins, so comments elsewhere can't change the header.
func parseHeader(newp *patch.Patch, data []byte) {
	seen := make(map[string]bool)

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			return
		}

		parts := headerLine.FindStringSubmatch(line)
		// parts[0] == whole line
		// parts[1] == keyword
		// parts[2] == value
		if len(parts) < 2 || seen[parts[1]] {
			continue
		}
		seen[parts[1]] = true

		val := strings.TrimSpace(parts[2])
		switch parts[1] {
		case "id":
			newp.Id = val
		case "prereqs":
			newp.Prereqs = splitList(val)
		case "description":
			newp.Description = val
		case "options":
			newp.Options = splitList(val)
		}
	}
}

// splitList splits a comma separated header value, dropping surrounding
// whitespace and empty items.
func splitList(val string) []string {
	var items []string
	for _, item := range strings.Split(val, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func (p *Patcher) NewPatch(thePath string) (*patch.Patch, error) {
	data, err := os.ReadFile(thePath)
	if err != nil {
		return nil, err
	}

	newp := patch.Patch{
		Filename: thePath,
		Body:     data,
	}
	parseHeader(&newp, data)

	if newp.Id == "" {
		// The ID field is an absolute must. Without it, there is
		// no linking.
		return nil, fmt.Errorf("file %q missing ID field", thePath)
	}

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
		sort.Strings(filenames)
		return fmt.Errorf("loop detected:\n%s", strings.Join(filenames, "\n"))
	}

	thepatch.Weight += 1
	detectionMap[thepatch.Id] = thepatch

	for _, patchkey := range thepatch.Prereqs {

		prereq, ok := p.patches[patchkey]
		if !ok {
			return fmt.Errorf("bad ID reference. File %q refers to id %q which doesn't exist", thepatch.Filename, patchkey)
		}

		err := p.bumpWeight(prereq, detectionMap)
		if err != nil {
			return err
		}
	}
	delete(detectionMap, thepatch.Id)
	return nil
}

func (p *Patcher) Resolve() error {
	ids := &set.Set{}
	var err error

	if !p.dry {
		ids, err = p.engine.GetInstalledIDs()
		if err != nil {
			return err
		}
	}

	p.installedIDs = ids
	var patches []*patch.Patch

	for _, thispatch := range p.patches {
		if p.installedIDs.Contains(thispatch.Id) {
			fmt.Println("Skipping already installed patch", thispatch.Id)
			continue
		}
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
	if p.verbose {
		fmt.Println("Patches to be applied:")
		for _, ptch := range patches {
			fmt.Printf("[weight %d] %s\n", ptch.Weight, ptch.Filename)
		}
	}
	p.ordered = patches
	return nil
}

func (p *Patcher) shouldIgnore(thePath string) bool {
	if p.ignore == "" {
		return false
	}

	parts := strings.Split(p.ignore, ",")
	for _, part := range parts {
		fullPath := filepath.Join(p.folder, part)
		if filepath.Clean(thePath) == fullPath {
			return true
		}
	}

	return false
}

func (p *Patcher) walkDirFunc(thePath string, d fs.DirEntry, err error) error {

	if p.shouldIgnore(thePath) {
		if p.verbose {
			fmt.Println("ignoring:", thePath)
		}
		if d != nil && d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

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
	numDone := 0

	// Make sure to apply the init_patch.sql file first
	if p.installedIDs.Len() == 0 && !p.dry {
		fmt.Println("no IDs detected, patcher not installed yet")
		if p.verbose || p.dry {
			fmt.Printf("applying: (weight %d) %s\n", p.initPatch.Weight, p.initPatch.Filename)
		}

		if err := p.engine.Patch(p.initPatch); err != nil {
			return err
		}
		// Skip applying this in the following loop
		if err := p.installedIDs.Add(p.initPatch.Id); err != nil {
			return err
		}
		numDone += 1
	}

	for _, thepatch := range p.ordered {
		if p.installedIDs.Contains(thepatch.Id) {
			if p.verbose {
				fmt.Printf("INSTALLED: (weight %d) %s\n", thepatch.Weight, thepatch.Filename)
			}
			continue
		}
		if p.verbose || p.dry {
			extra := ""
			if p.dry {
				extra = " [dry]"
			}
			fmt.Printf("applying: (weight %d) %s%s\n", thepatch.Weight, thepatch.Filename, extra)
		}

		if !p.dry {
			if err := p.engine.Patch(thepatch); err != nil {
				return err
			}
		}
		numDone += 1
	}
	if numDone > 0 {
		patchstr := "patches"
		if numDone == 1 {
			patchstr = "patch"
		}
		fmt.Printf("Applied %d %s successfully\n", numDone, patchstr)
	} else {
		fmt.Println("No work to do")
	}

	return nil
}
