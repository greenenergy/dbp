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
package dbe

import (
	"fmt"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
)

type MockDBE struct {
}

func NewMockDBE() DBEngine {
	return &MockDBE{}
}

func (p *MockDBE) GetInstalledIDs() (*set.Set, error) {
	return &set.Set{}, nil
}

func (p *MockDBE) Patch(thepatch *patch.Patch) error {
	fmt.Println(thepatch.Filename)
	return nil
}
