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
package patch

import "strings"

type Patch struct {
	Id          string   `json:"id"`
	Description string   `json:"description"`
	Prereqs     []string `json:"prereqs"`
	Body        []byte   `json:"-"`
	Weight      int      `json:"weight"`
	Filename    string   `json:"filename"`
	Options     []string `json:"options"`
}

func (p *Patch) HasOption(o string) bool {
	for _, option := range p.Options {
		// Case insensitive compare
		if strings.EqualFold(option, o) {
			return true
		}
	}
	return false
}

type ByWeight []*Patch

func (by ByWeight) Len() int {
	return len(by)
}

func (by ByWeight) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

func (by ByWeight) Less(i, j int) bool {
	if by[i].Weight == by[j].Weight {
		return by[i].Id < by[j].Id
	}
	return by[i].Weight > by[j].Weight
}
