package dbe

import (
	"fmt"

	"github.com/greenenergy/migrate/pkg/patch"
)

type PGDBE struct {
}

func NewPGDBE() DBEngine {
	return &PGDBE{}
}

func (p *PGDBE) IsConfigured() bool {
	return false
}

func (p *PGDBE) Configure() error {
	return fmt.Errorf("unimplemented")
}

func (p *PGDBE) GetInstalledIDs() ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *PGDBE) Patch(*patch.Patch) error {
	return fmt.Errorf("unimplemented")
}
