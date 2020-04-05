/*! \file main.go
	\brief Main cockroach stuff
*/

package cockroach 

import (
	
	"github.com/pkg/errors"

	"database/sql"
	"context"
	"time"
	"encoding/json"
)

// if errors.Cause (err) != sql.ErrNoRows { //not an error we can ignore
var db *sql.DB // global sql database handle

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DATABASE FUNCTIONS ------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func SetDB (inDB *sql.DB) error {
	db = inDB
	return TestDB() // make sure we can ping this
}

func TestDB () error {
	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)
	defer cancel()
	
	return db.PingContext(ctx)
}

// all cockroach classes are build upon this
type toolz_c struct {
	
}

func (this *toolz_c) RowsChk (prevErr error, rows *sql.Rows) error {
	rows.Close() //close it here, no harm in doing it more than once
	return errors.WithStack (rows.Err())
}

func (this *toolz_c) Exec (query string, args ...interface{}) error {
	_, err := db.Exec (query, args...)
	return errors.WithStack (err)
}

func (this *toolz_c) GenUUID () (out string) {
	db.QueryRow(`SELECT gen_random_uuid()`).Scan(&out)
	return
}

/*! \brief Tries to unmarshal an object and records an error if it happens
*/
func (this *toolz_c) UM (jAttr []byte, out interface{}) error {
    err := json.Unmarshal(jAttr, out)
    if err != nil { //this didn't work
        if len(jAttr) >= 2 {
			return errors.Wrapf(err, "byte string : %s", string(jAttr))
        } else {
			return errors.Wrap(err, "json input is empty")
        }
    }
    return nil
}
