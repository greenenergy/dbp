package dbe

import (
	"errors"

	"github.com/greenenergy/migrate/pkg/patch"
	"github.com/spf13/pflag"
)

type PGDBE struct {
}

func NewPGDBE(flags *pflag.FlagSet) DBEngine {
	return &PGDBE{}
}

func (p *PGDBE) IsConfigured() bool {
	return false
}

func (p *PGDBE) Configure() error {
	return errors.New("postgres engine: Configure() unimplemented")
}

func (p *PGDBE) GetInstalledIDs() ([]string, error) {
	return nil, errors.New("postgres engine: GetInstalledIDs() unimplemented")
}

func (p *PGDBE) Patch(*patch.Patch) error {
	return errors.New("postgres engine: Patch() unimplemented")
}
