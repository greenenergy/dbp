package dbe

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/jmoiron/sqlx"
)

type MySQLDBE struct {
	conn    *sqlx.DB
	verbose bool
	debug   bool
	retries int
}

type MySqlArgs struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// func NewMySQLDBE(host string, port int, user, password, dbname string, sslmode bool, verbose, debug bool, retries int) (DBEngine, error) {
func NewMySQLDBE(args *EngineArgs) (DBEngine, error) {
	//connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
	//	args.Host, args.Port, args.Username, args.Name, mode, args.Password)

	root := "%s:%s@tcp(%s:%d)/%s?multiStatements=true"
	connStr := fmt.Sprintf(root, args.Username, args.Password, args.Host, args.Port, args.Name)

	safeConnStr := fmt.Sprintf(root, args.Username, "<redacted>", args.Host, args.Port, args.Name)

	if args.Verbose {
		fmt.Println("connstr:", safeConnStr)
	}

	conn, err := sqlx.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error opening database:", err.Error())
		return nil, fmt.Errorf("couldn't open postgres: %v", err.Error())
	}

	// Once we have a connection, we should check for our table & data.
	// If it doesn't exist, then this is the first time we're running
	// here. If it does exist, then we should read those IDs in as well
	// and seed the memory map of IDs. The question here is -- do we expect
	// all the patch files to be present with a given database? If we do
	// expect that, then we wouldn't need to load in the existing patches.

	mysqldbe := MySQLDBE{
		conn:    conn,
		verbose: args.Verbose,
		debug:   args.Debug,
		retries: args.Retries,
	}

	err = mysqldbe.checkInstall()
	if err != nil {
		return nil, err
	}

	return &mysqldbe, nil
}

func (p *MySQLDBE) checkInstall() error {
	success := false

	p.DPrint("checkInstall()...")

	for x := 0; x < p.retries; x++ {
		p.DPrint("querying dbp_patch_table...")

		_, err := p.conn.Queryx("select count(*) from dbp_patch_table")
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				fmt.Printf("this is a network error(%q), retrying %d more times\n", err.Error(), p.retries-x)
				time.Sleep(time.Second)
			} else {
				p.DPrint("creating dbp_patch_table...")

				_, err = p.conn.Queryx(`
create table dbp_patch_table (
	id text,
	created timestamp not null default CURRENT_TIMESTAMP,
	prereqs varchar(256),
	description varchar(256),
 	primary key (id(255))
)
`)
				if err != nil {
					return fmt.Errorf("problem creating patch table:%q", err)
				}
				success = true
				break
			}
		} else {
			success = true
			break
		}
	}
	if !success {
		fmt.Println("couldn't find the db")
		log.Fatal("no database")
	}
	return nil
}

func (p *MySQLDBE) GetInstalledIDs() (*set.Set, error) {
	q, err := p.conn.Query("select id from dbp_patch_table")
	if err != nil {
		return nil, fmt.Errorf("problem getting installed ids: %s", err.Error())
	}
	defer q.Close()

	output := set.Set{}

	for q.Next() {
		var id string
		err = q.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("problem scanning id: %s", err.Error())
		}
		output.Add(id)
	}
	return &output, nil
}

func (p *MySQLDBE) DPrint(msg string) {
	if p.debug {
		fmt.Println(msg)
	}
}

func (p *MySQLDBE) Patch(ptch *patch.Patch) error {
	fmt.Println("Begin transaction, patching:", ptch.Filename)

	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("problem while trying to apply patch: %s", err.Error())
	}

	_, err = tx.Exec(string(ptch.Body))
	if err != nil {
		fmt.Println("rolling back due to error:", err.Error())
		err2 := tx.Rollback()
		if err2 != nil {
			fmt.Println("error rolling back:", err2.Error())
		} else {
			fmt.Println("NO ERROR ROLLING BACK!")
		}
		return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
	}

	prereqs := strings.Join(ptch.Prereqs, ",")
	_, err = tx.Exec("insert into dbp_patch_table(id, prereqs, description) values (?, ?, ?)",
		ptch.Id, prereqs, ptch.Description)

	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "unexpected Parse response") {
			return fmt.Errorf("problem applying patch file %q: %s -- Are you returning a resultset, or possibly have a commit statement?", ptch.Filename, err.Error())
		}
		return fmt.Errorf("problem applying patch file %q: %s", ptch.Filename, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("*** Problem committing:", err.Error())
	}
	return nil
}
