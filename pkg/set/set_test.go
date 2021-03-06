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

import (
	"testing"
)

func TestUnion(t *testing.T) {
	var s1, s2 Set

	s1.Add("hello")
	s2.Add("hello", "world")

	err := s1.Union(&s2)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if s1.Len() != s2.Len() {
		t.Fatalf("wrong set length. Should be %d, was %d", s2.Len(), s1.Len())
	}
}
