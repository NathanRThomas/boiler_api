/*! \file box.go
	\brief Contains re-used functions regargding the cockroach database "stuff"
*/

package cockroach 

import (
	"github.com/NathanRThomas/boiler_api/pkg/models"
	
	"github.com/pkg/errors"

	"database/sql"
	"context"
	"time"
	"encoding/json"
	"log"
	"runtime"
	"strings"
	"regexp"
)

// if errors.Cause (err) != sql.ErrNoRows { //not an error we can ignore

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DATABASE FUNCTIONS ------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type DB_c struct {
	DB *sql.DB
}

/*! \brief All our context threads get their own transaction to apply changes to
*/
func (this *DB_c) Begin (ctx context.Context) (context.Context, *sql.Tx, error) {
	tx, err := this.DB.BeginTx (ctx, nil)
	if err == nil {
		ctx = context.WithValue(ctx, "tx", tx)
		ctx = context.WithValue(ctx, "startTime", time.Now())
	}
	return ctx, tx, err
}

/*! \brief Handles committing or rolling back based on the passed in error
*/
func (this *DB_c) Finish (ctx context.Context, tx *sql.Tx, err error) error {
	if err == nil { // nothing wrong, go for it
		err = errors.WithStack (tx.Commit())
	} else if errors.Cause (err) == models.ErrType_nonFatal {	// this is a non-fatal error, so we still want to commit the transaction, and report this error up
		lErr := errors.WithStack (tx.Commit())
		if lErr != nil {	// we had an error committing
			err = errors.Wrap (lErr, err.Error())
		}
	} else {
		if lErr := tx.Rollback(); lErr != nil {	// this is bad, so rollback this transaction
			err = errors.Wrap (lErr, err.Error())
		}
	}

	if err == nil {	// only worry about the timming if we don't have other errors to worry about
		startTime, ok := ctx.Value ("startTime").(time.Time)
		if ok {
			if time.Now().Sub(startTime).Milliseconds() > 4000 {
				err = errors.Wrapf (models.ErrType_tookLongTime, "milliseconds: %d", time.Now().Sub(startTime).Milliseconds())
			}
		} else {
			err = errors.New("no startTime found in the context")
		}
	}
	
	return err
}

func (this *DB_c) Test () error {
	var ctx context.Context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return this.DB.PingContext(ctx)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- RE-USED FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type toolz_c struct {
	
}

func (this *toolz_c) RowsChk (prevErr error, rows *sql.Rows) error {
	rows.Close() //close it here, no harm in doing it more than once
	err := errors.WithStack (rows.Err())
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			log.Println (file, no, err)
		} else {
			log.Println (err)
		}

		if prevErr == nil {
			return err
		} else {
			return errors.Wrap (prevErr, err.Error())	// we're coming in with an error, so wrap it
		}
	}
	return prevErr
}

func (this *toolz_c) GenUUID (tx *sql.Tx) (out string) {
	tx.QueryRow(`SELECT gen_random_uuid()`).Scan(&out)
	return
}

/*! \brief Returns a nullable object depending on if we liked our input string
*/
func (this *toolz_c) Nullable (in string) (out sql.NullString) {
	if len(in) > 0 {
		out.String = in 
		out.Valid = true
	}
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

func (this *toolz_c) StrLimit (in string, limit int) string {
	in = strings.TrimSpace(in)
	if len(in) > limit {
		in = in[:limit]
	}
	return in
}

/*! \brief Preps a string for use in the database by removing "weird" characters
 */
 func (this *toolz_c) safeRegex (in string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9]")
	return strings.ToLower(r.ReplaceAllString(in, ""))
}

/*! \brief This will convert a raw user inputed search string and make it into an array of "useful" results
  This is also what we use to convert the user names from facebook into something searchable
*/
func (this *toolz_c) SearchTerms (in string) []string {
	out := make([]string, 0)
	loc := strings.Trim(strings.Replace(in, "-", " ", -1), " ")
	for _, str := range strings.Fields(loc) {
		l := this.safeRegex(str)
		if len(l) > 1 {
			out = append(out, l)
		}
	}
	return out
}

/*! \brief Joins an array of uuid's in a string for using an "in" query
*/
func (this *toolz_c) Join (in []models.UUID) string {
	str := make([]string, 0)

	for _, i := range in {
		str = append (str, i.String())
	}

	return "'" + strings.Join(str, "','") + "'"
}
