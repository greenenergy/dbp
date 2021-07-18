package dbe

import (
	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
)

type DBEngine interface {
	IsConfigured() bool                 // Has this database already been used by migrate?
	Configure() error                   // Set up the table needed to track patches
	GetInstalledIDs() (*set.Set, error) // Return the IDs of patches that have already been installed
	Patch(*patch.Patch) error
}
