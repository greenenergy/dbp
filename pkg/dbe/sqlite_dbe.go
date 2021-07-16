package dbe

import (
	"errors"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/spf13/pflag"
)

type SQLiteDBE struct {
}

func NewSQLiteDBE(flags *pflag.FlagSet) DBEngine {
	return &SQLiteDBE{}
}

func (p *SQLiteDBE) IsConfigured() bool {
	return false
}

func (p *SQLiteDBE) Configure() error {
	return errors.New("sqlite engine: Configure() unimplemented")
}

func (p *SQLiteDBE) GetInstalledIDs() ([]string, error) {
	return nil, errors.New("sqlite engine: GetInstalledIDs() unimplemented")
}

func (p *SQLiteDBE) Patch(*patch.Patch) error {
	return errors.New("sqlite engine: Patch() unimplemented")
}
