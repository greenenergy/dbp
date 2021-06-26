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

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new patch file",
	Long: `Helper function to create a new patch file for you.
It just fills out the header parameters and generates an ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("-- PATCH: v0.0.1")
		fmt.Println("-- id:", uuid.New().String())
		fmt.Println("-- author: ")
		fmt.Println("-- prereqs: ")
		fmt.Println("-- description: ")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
