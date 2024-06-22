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
	"reflect"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
)

type DBEngine interface {
	GetInstalledIDs() (*set.Set, error) // Return the IDs of patches that have already been installed
	Patch(*patch.Patch) error
}

type EngineArgs struct {
	Host        string `json:"host" connarg:"host"`
	Port        int    `json:"port" connarg:"port"`
	Name        string `json:"dbname" connarg:"dbname"`
	Username    string `json:"user" connarg:"user"`
	Password    string `json:"password" connarg:"password,protected"`
	SSLMode     string `json:"sslmode" connarg:"sslmode"`
	SSLCert     string `json:"sslcert" connarg:"sslcert"`
	SSLKey      string `json:"sslkey" connarg:"sslkey"`
	SSLRootCert string `json:"sslrootcert" connarg:"sslrootcert"`
	Debug       bool   `json:"debug"`
	Verbose     bool   `json:"verbose"`
	Retries     int    `json:"retries"`
}

func NewEngineArgs(flags *flag.FlagSet) *EngineArgs {
	ea := &EngineArgs{}
	ea.AddFlags(flags)
	return ea
}

func (ea *EngineArgs) AddFlags(flags *flag.FlagSet) {
	flags.BoolVarP(&ea.Verbose, "verbose", "v", false, "be verbose")
	flags.IntVarP(&ea.Retries, "retries", "r", 10, "Number of retries when trying to connect")
	flags.StringVarP(&ea.Host, "db.host", "", "", "hostname of db server")
	flags.IntVarP(&ea.Port, "db.port", "", 5432, "Port to connect to")
	flags.StringVarP(&ea.Username, "db.username", "", "", "Username to use for db")
	flags.StringVarP(&ea.Password, "db.password", "", "", "Password to use for db")
	flags.StringVarP(&ea.Name, "db.name", "", "", "database name")
	flags.StringVarP(&ea.SSLMode, "db.sslmode", "", "require", "ssl mode to use. Options are: disable, allow, prefer, require, verify-ca and verify-full. See https://www.postgresql.org/docs/current/libpq-ssl.html for details")
	flags.StringVarP(&ea.SSLCert, "db.sslcert", "", "", "path to cert")
	flags.StringVarP(&ea.SSLKey, "db.sslkey", "", "", "path to key")
	flags.StringVarP(&ea.SSLRootCert, "db.sslrootcert", "", "", "path to verification root cert")
}

func (ea *EngineArgs) ToConnStr(protectPassword bool) string {
	var resultlist []string
	result := ""
	otype := reflect.TypeOf(ea)
	oval := reflect.ValueOf(ea)
	if oval.Kind() == reflect.Ptr {
		oval = oval.Elem()
		otype = otype.Elem()
	}
	for i := 0; i < otype.NumField(); i++ {
		field := otype.Field(i)

		connarg := field.Tag.Get("connarg")
		if connarg == "" {
			continue
		}
		parts := strings.Split(connarg, ",")
		fieldname := parts[0]
		fval := oval.Field(i)
		val := fval.Interface()
		switch t := val.(type) {
		case string:
			if t == "" {
				// Don't add empty args
				continue
			}
			if len(parts) == 2 {
				if protectPassword {
					t = "<redacted>"
				}
			}
			resultlist = append(resultlist, fmt.Sprintf("%s=%s", fieldname, t))
		case int:
			resultlist = append(resultlist, fmt.Sprintf("%s=%d", fieldname, t))
		default:
			fmt.Println("unsupported field type:", t)
		}
	}
	result = strings.Join(resultlist, " ")
	return result
}
