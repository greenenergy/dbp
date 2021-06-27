package dbe

import (
	"fmt"

	"github.com/greenenergy/migrate/pkg/patch"
)

type SQLiteDBE struct {
}

func NewSQLiteDBE() DBEngine {
	return &SQLiteDBE{}
}

func (p *SQLiteDBE) IsConfigured() bool {
	return false
}

func (p *SQLiteDBE) Configure() error {
	return fmt.Errorf("unimplemented")
}

func (p *SQLiteDBE) GetInstalledIDs() ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *SQLiteDBE) Patch(*patch.Patch) error {
	return fmt.Errorf("unimplemented")
}
