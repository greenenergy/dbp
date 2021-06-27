package dbe

import (
	"fmt"

	"github.com/greenenergy/migrate/pkg/patch"
)

type MEMDBE struct {
}

func NewMEMDBE() DBEngine {
	return &MEMDBE{}
}

func (p *MEMDBE) IsConfigured() bool {
	return false
}

func (p *MEMDBE) Configure() error {
	return fmt.Errorf("unimplemented")
}

func (p *MEMDBE) GetInstalledIDs() ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *MEMDBE) Patch(*patch.Patch) error {
	return fmt.Errorf("unimplemented")
}
