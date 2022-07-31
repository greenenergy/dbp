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

	engine := dbe.NewMockDBE()

	p, err := NewPatcher(false, true, engine)
	if err != nil {
		t.Fatal(err)
	}
	errorCases := []struct {
		folder string
		err    error
	}{
		{
			"testdata/dupe_id",
			fmt.Errorf(`duplicate id: fece2b8e-cf43-11eb-b7f3-07af1b70a47a, found in File "testdata/dupe_id/init_patch.sql" and "testdata/dupe_id/patch_1.sql"`),
		},
		{
			"testdata/long_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/long_loop/loop_a.sql\n" +
				"testdata/long_loop/loop_c.sql\n" +
				"testdata/long_loop/loop_f.sql\n" +
				"testdata/long_loop/loop_g.sql"),
		},
		{
			"testdata/missing_id_1",
			fmt.Errorf(`file "testdata/missing_id_1/init_patch.sql" missing ID field`),
		},
		{
			"testdata/missing_id_2",
			fmt.Errorf(`bad ID reference. File "testdata/missing_id_2/a.sql" refers to id "this-is-not-a-valid-id" which doesn't exist`),
		},
		{
			"testdata/shortest_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/shortest_loop/init_patch.sql"),
		},
		{
			"testdata/short_loop",
			fmt.Errorf("loop detected:\n" +
				"testdata/short_loop/feat_abcd.sql\n" +
				"testdata/short_loop/feat_efg.sql"),
		},
	}

	for _, c := range errorCases {
		fmt.Println("Scanning:", c.folder)
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
}
