/*! \file db.go
    \brief Main database class/package for communication with databases

*/

package db

import (
	"database/sql"
	"fmt"
    "log"
    _ "github.com/lib/pq"
	
	"github.com/NathanRThomas/boiler_api/toolz"
)
var sqlDB *sql.DB    //this is the current "active" connection

const max_open_pg_conns	= 500

type cockroach_c struct {
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *cockroach_c) connect (ip string, port int, database string) (*sql.DB, error) {
    sslmode := "sslmode=disable"
    if toolz.AppConfig.ProductionFlag == toolz.ProductionType_Production {
        sslmode = "sslmode=verify-full&sslcert=/var/certs/client.root.crt&sslkey=/var/certs/client.root.key&sslrootcert=/var/certs/ca.crt"
    }
    locDB, err := sql.Open("postgres", fmt.Sprintf("postgres://root@%s:%d/%s?%s",
                                                  ip, port, database, sslmode))

    if err == nil {
        err = this.test(locDB)
        if err == nil {
            locDB.SetMaxOpenConns(max_open_pg_conns)
            locDB.Exec(`SET TIME ZONE 'UTC'`)   //set our default timezone
        }
    }
    return locDB, ErrChk(err)
}

/*! \brief Simple query to ensure we have a valid connection to the database
*/
func (this *cockroach_c) test(locDB *sql.DB) (err error) {
    res := 0
    err = ErrChk(locDB.QueryRow("SELECT 2 + 2").Scan(&res))
    if res != 4 { return ErrChk(fmt.Errorf("DB Error, things don't add up: %d", res)) }
    return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- INIT FUNCTIONS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *cockroach_c) Init (running *bool) {
	if len(toolz.AppConfig.Cockroach.IP) == 0 { return }	//nothing to do

	locDB, err := this.connect(toolz.AppConfig.Cockroach.IP, toolz.AppConfig.Cockroach.Port, toolz.AppConfig.Cockroach.Database)
	if err == nil {
		sqlDB = locDB
	} else {
		ErrChk(err)  //this is bad
		log.Panic("All cockroach database connections failed!")
	}
    return //we're good
}

/*! \brief Simple query to ensure we have a valid connection to the database
*/
func (this *cockroach_c) Test() (err error) {
    return this.test(sqlDB)
}

/*! \brief Closes down the connection pool when we shutdown the service
*/
func (this *cockroach_c) Close () {
	if sqlDB != nil {
		sqlDB.Close()
	}
}

var cockroach cockroach_c

func dbExec (query string, args ...interface{}) error {
    _, err := sqlDB.Exec(query, args...)
    return ErrChk(err)
}

func TestCockroach () error {
	return cockroach.Test()
}