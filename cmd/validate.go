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

	"github.com/greenenergy/dbp/pkg/dbe"
	"github.com/greenenergy/dbp/pkg/patcher"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "make sure there are no dangling references, or any other issues",
	Long: `Make sure there are no problems with the folder of patches.
Problems could include multiple files with the same ID, referring to
an ID that doesn't exist, etc`,
	Run: func(cmd *cobra.Command, args []string) {
		engine := dbe.NewMockDBE()

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatal(err)
		}

		folder, err := cmd.Flags().GetString("folder")
		if err != nil {
			log.Fatal(err.Error())
		}
		if folder == "" {
			log.Fatal("you must specify a folder")
		}

		ignore, err := cmd.Flags().GetString("ignore")
		if err != nil {
			log.Fatal(err)
		}

		p, err := patcher.NewPatcher(false, verbose, engine, folder, ignore)
		if err != nil {
			log.Fatal(err)
		}
		p.Dry(true)

		err = p.Scan(folder)
		if err != nil {
			fmt.Println("error scanning:", err)
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringP("folder", "f", "", "set the processing folder")
}
