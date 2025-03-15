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
	"testing"

	"github.com/greenenergy/dbp/pkg/dbe"
)

func TestPatcher(t *testing.T) {
	t.Skip("temp")
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
		if err != nil {
			if c.err == nil {
				t.Errorf("error scanning: %v", err)
			} else if err.Error() != c.err.Error() {
				t.Errorf("error scanning. Expected:\n%v\ngot:\n%v", c.err, err)
			}
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

func TestIgnore(t *testing.T) {
	engine := dbe.NewMockDBE()

	p, err := NewPatcher(false, true, engine, "", "")
	if err != nil {
		t.Fatal(err)
	}
	p.ignore = "/patchset_1"

	successCases := []string{"./testdata/good"}

	for _, c := range successCases {
		p.Reset()
		err = p.Scan(c)
		if err != nil {
			t.Errorf("error scanning: %v", err)
		}
		//err = p.Process()
		//if err != nil {
		//	t.Fatal(err)
		//}
	}

}
