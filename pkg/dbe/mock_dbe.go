package dbe

import (
	"fmt"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/spf13/pflag"
)

type MockDBE struct {
}

func NewMockDBE(flags *pflag.FlagSet) DBEngine {
	return &MockDBE{}
}

func (p *MockDBE) IsConfigured() bool {
	return false
}

func (p *MockDBE) Configure() error {
	return fmt.Errorf("unimplemented")
}

func (p *MockDBE) GetInstalledIDs() (*set.Set, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *MockDBE) Patch(thepatch *patch.Patch) error {
	fmt.Println(thepatch.Filename)
	return nil
}
