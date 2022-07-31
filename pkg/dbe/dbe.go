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
	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
)

type DBEngine interface {
	GetInstalledIDs() (*set.Set, error) // Return the IDs of patches that have already been installed
	Patch(*patch.Patch) error
}

type EngineArgs struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
	Debug    bool   `json:"debug"`
	Verbose  bool   `json:"verbose"`
	Retries  int    `json:"retries"`
}
