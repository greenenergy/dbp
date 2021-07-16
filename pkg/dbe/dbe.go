package dbe

import "github.com/greenenergy/dbp/pkg/patch"

type DBEngine interface {
	IsConfigured() bool                 // Has this database already been used by migrate?
	Configure() error                   // Set up the table needed to track patches
	GetInstalledIDs() ([]string, error) // Return the IDs of patches that have already been installed
	Patch(*patch.Patch) error
}
