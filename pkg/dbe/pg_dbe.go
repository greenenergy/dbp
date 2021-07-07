package dbe

import (
	"errors"
	"fmt"

	"github.com/greenenergy/migrate/pkg/patch"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // This  is the postgres driver for sqlx
	"github.com/spf13/pflag"
)

type PGDBE struct {
	conn *sqlx.DB
}

func NewPGDBE(flags *pflag.FlagSet) (DBEngine, error) {

	var hostname, user, dbname, dbpass, port string

	if flags != nil {
		hostname = flags.Lookup("dbhost").Value.String()
		dbname = flags.Lookup("dbname").Value.String()
		dbpass = flags.Lookup("dbpass").Value.String()
		port = flags.Lookup("port").Value.String()
	}

	mode := "disable"

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		hostname, port, user, dbname, mode, dbpass)

	conn, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't open postgres: %v", err.Error())
	}
	return &PGDBE{conn: conn}, nil
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
