package util

import (
	"fmt"

	"github.com/greenenergy/dbp/pkg/dbe"
	"github.com/spf13/pflag"
)

func FlagsToArgs(flags *pflag.FlagSet) (*dbe.EngineArgs, error) {
	var args dbe.EngineArgs
	var err error

	args.Host, err = flags.GetString("db.host")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Port, err = flags.GetInt("db.port")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Username, err = flags.GetString("db.username")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Password, err = flags.GetString("db.password")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Name, err = flags.GetString("db.name")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.SSL, err = flags.GetBool("db.tls")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Verbose = flags.Lookup("verbose").Value.String() == "true"

	args.Debug, err = flags.GetBool("debug")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	args.Retries, err = flags.GetInt("retries")
	if err != nil {
		return nil, fmt.Errorf("problem reading flag: %s", err.Error())
	}

	return &args, nil
}
