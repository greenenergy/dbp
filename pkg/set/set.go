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
package set

import "fmt"

// --------------------------

type Set struct {
	init     bool
	contents map[string]bool
}

func NewSet() *Set {
	return &Set{
		init:     true,
		contents: make(map[string]bool),
	}
}

func (s *Set) Contains(key string) bool {
	if !s.init {
		s.contents = make(map[string]bool)
		s.init = true
	}
	_, ok := s.contents[key]
	return ok
}

func (s *Set) Add(keys ...string) error {
	if !s.init {
		s.contents = make(map[string]bool)
		s.init = true
	}
	for _, key := range keys {
		if _, ok := s.contents[key]; ok {
			return fmt.Errorf("key exists")
		}
		s.contents[key] = true
	}
	return nil
}

func (s *Set) Union(s2 *Set) error {
	for key := range s2.contents {
		s.Add(key)
	}
	return nil
}
