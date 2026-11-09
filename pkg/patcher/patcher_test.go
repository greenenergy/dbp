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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/greenenergy/dbp/pkg/dbe"
	"github.com/greenenergy/dbp/pkg/patch"
)

func TestPatcher(t *testing.T) {
	engine := dbe.NewMockDBE()

	p, err := NewPatcher(false, true, engine, "", "")
	if err != nil {
		t.Fatal(err)
	}
	errorCases := []struct {
		folder string
		err    error
	}{
		{
			"testdata/bad/dupe_id",
			fmt.Errorf(`duplicate id: fece2b8e-cf43-11eb-b7f3-07af1b70a47a, found in File "testdata/bad/dupe_id/init_patch.sql" and "testdata/bad/dupe_id/patch_1.sql"`),
		},
		{
			"testdata/bad/long_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/bad/long_loop/loop_a.sql\n" +
				"testdata/bad/long_loop/loop_c.sql\n" +
				"testdata/bad/long_loop/loop_f.sql\n" +
				"testdata/bad/long_loop/loop_g.sql"),
		},
		{
			"testdata/bad/missing_id_1",
			fmt.Errorf(`file "testdata/bad/missing_id_1/init_patch.sql" missing ID field`),
		},
		{
			"testdata/bad/missing_id_2",
			fmt.Errorf(`bad ID reference. File "testdata/bad/missing_id_2/a.sql" refers to id "this-is-not-a-valid-id" which doesn't exist`),
		},
		{
			"testdata/bad/shortest_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/bad/shortest_loop/init_patch.sql"),
		},
		{
			"testdata/bad/short_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/bad/short_loop/feat_abcd.sql\n" +
				"testdata/bad/short_loop/feat_efg.sql"),
		},
	}

	for _, c := range errorCases {
		p.Reset()
		err = p.Scan(c.folder)
		if err == nil {
			t.Errorf("scanning %s: expected an error", c.folder)
		} else if err.Error() != c.err.Error() {
			t.Errorf("error scanning. Expected:\n%v\ngot:\n%v", c.err, err)
		}
	}

	successCases := []string{"testdata/good/patchset_1"}
	for _, c := range successCases {
		p.Reset()
		err = p.Scan(c)
		if err != nil {
			t.Errorf("error scanning: %v", err)
		}
		err = p.Process()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// writePatches creates a patch tree in a temp folder from a map of relative
// path to file contents, and returns the folder.
func writePatches(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func scan(t *testing.T, folder, ignore string) (*Patcher, error) {
	t.Helper()
	p, err := NewPatcher(true, false, dbe.NewMockDBE(), folder, ignore)
	if err != nil {
		t.Fatal(err)
	}
	return p, p.Scan(folder)
}

const initPatch = "-- id: init\nselect 1;\n"

func TestIgnore(t *testing.T) {
	dir := writePatches(t, map[string]string{
		"init_patch.sql":    initPatch,
		"keep.sql":          "-- id: keep\n-- prereqs: init\nselect 1;\n",
		"skipme/a.sql":      "-- id: dir-a\n-- prereqs: init\nselect 1;\n",
		"skipme/deep/b.sql": "-- id: dir-b\n-- prereqs: init\nselect 1;\n",
		"file.sql":          "-- id: single-file\n-- prereqs: init\nselect 1;\n",
		"sub/skipme/c.sql":  "-- id: same-name-elsewhere\n-- prereqs: init\nselect 1;\n",
	})

	p, err := scan(t, dir, "skipme,file.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"init", "keep", "same-name-elsewhere"} {
		if _, ok := p.patches[id]; !ok {
			t.Errorf("patch %q should have been scanned", id)
		}
	}
	for _, id := range []string{"dir-a", "dir-b", "single-file"} {
		if _, ok := p.patches[id]; ok {
			t.Errorf("patch %q should have been ignored", id)
		}
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		id      string
		prereqs []string
		options []string
	}{
		{
			name:    "all fields",
			body:    "-- id: a\n-- prereqs: b,c\n-- options: chop\n\nselect 1;\n",
			id:      "a",
			prereqs: []string{"b", "c"},
			options: []string{"chop"},
		},
		{
			name:    "spaces around list items",
			body:    "-- id: a\n-- prereqs: b, c ,d\n-- options: other, chop\nselect 1;\n",
			id:      "a",
			prereqs: []string{"b", "c", "d"},
			options: []string{"other", "chop"},
		},
		{
			name:    "CRLF line endings",
			body:    "-- id: a\r\n-- prereqs: b\r\nselect 1;\r\n",
			id:      "a",
			prereqs: []string{"b"},
		},
		{
			name: "blank prereqs",
			body: "-- id: a\n-- prereqs:  \nselect 1;\n",
			id:   "a",
		},
		{
			name:    "comments and blanks before the header",
			body:    "\n-- A note about this patch\n\n-- id: a\n-- prereqs: b\nselect 1;\n",
			id:      "a",
			prereqs: []string{"b"},
		},
		{
			name:    "first occurrence wins in the leading comments",
			body:    "-- id: a\n-- prereqs: b\n\n-- prereqs: see ticket 12\n-- id: other\nselect 1;\n",
			id:      "a",
			prereqs: []string{"b"},
		},
		{
			name:    "comments after SQL are not header",
			body:    "-- id: a\nselect 1;\n-- id: other\n-- prereqs: b\n-- options: chop\nselect 2;\n",
			id:      "a",
			prereqs: nil,
			options: nil,
		},
		{
			name: "id after SQL is not found",
			body: "select 1;\n-- id: a\n",
			id:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got patch.Patch
			parseHeader(&got, []byte(tt.body))
			if got.Id != tt.id {
				t.Errorf("id: got %q, want %q", got.Id, tt.id)
			}
			if !reflect.DeepEqual(got.Prereqs, tt.prereqs) {
				t.Errorf("prereqs: got %q, want %q", got.Prereqs, tt.prereqs)
			}
			if !reflect.DeepEqual(got.Options, tt.options) {
				t.Errorf("options: got %q, want %q", got.Options, tt.options)
			}
		})
	}
}

func TestNestedMissingPrereq(t *testing.T) {
	dir := writePatches(t, map[string]string{
		"init_patch.sql": initPatch,
		"a.sql":          "-- id: a\n-- prereqs: b\nselect 1;\n",
		"b.sql":          "-- id: b\n-- prereqs: gone\nselect 1;\n",
	})

	// Scan order depends on map iteration, so run it enough times to hit
	// the path through bumpWeight.
	for i := 0; i < 50; i++ {
		_, err := scan(t, dir, "")
		if err == nil || !strings.Contains(err.Error(), `refers to id "gone"`) {
			t.Fatalf("expected a bad ID reference error, got: %v", err)
		}
	}
}
