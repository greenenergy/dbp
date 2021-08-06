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
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/greenenergy/dbp/pkg/patcher"
	"github.com/spf13/cobra"
)

// dbCmd represents the db command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply patches to a database",
	Long: `This function will collect all the patch files under a given folder,
order them according to prerequisites and apply them to an indicated database.
The filenames are not used by the patching system, or the directory tree, so you
are free to use them however you wish to organize your data.`,
	Run: func(cmd *cobra.Command, args []string) {
		folder := cmd.Flags().Lookup("folder").Value.String()

		p, err := patcher.NewPatcher(cmd.Flags())
		if err != nil {
			log.Fatal(err)
		}

		err = p.Scan(folder)
		if err != nil {
			fmt.Println("error scanning:", err)
			return
		}
		err = p.Process()
		if err != nil {
			fmt.Println("Problem applying patches:", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("folder", "f", "", "set the processing folder")
	applyCmd.Flags().BoolP("verbose", "v", false, "be verbose")
}
