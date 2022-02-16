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
	initialized bool
	contents    map[string]bool
}

func (s *Set) init() {
	if !s.initialized {
		s.contents = make(map[string]bool)
		s.initialized = true
	}
}

func NewSet() *Set {
	return &Set{
		initialized: true,
		contents:    make(map[string]bool),
	}
}
func (s *Set) Len() int {
	return len(s.contents)
}
func (s *Set) Contains(key string) bool {
	s.init()
	_, ok := s.contents[key]
	return ok
}

func (s *Set) Add(keys ...string) error {
	s.init()
	for _, key := range keys {
		if _, ok := s.contents[key]; ok {
			return fmt.Errorf("key %q exists", key)
		}
		s.contents[key] = true
	}
	return nil
}

func (s *Set) Union(s2 *Set) error {
	s.init()
	for key := range s2.contents {
		if !s.Contains(key) {
			err := s.Add(key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
