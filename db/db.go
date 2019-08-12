/*! \file db.go
    \brief Main database class/package for communication with databases

*/

package db

import (
	//"fmt"
	"database/sql"
	"regexp"
	"strings"
	"time"

	"github.com/NathanRThomas/boiler_api/toolz"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTANTS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const salt = "super salty" //salt for passwords

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Handles connections to all our databases
 */
func ConnectEm(running *bool) {
	cockroach.Init(running) //cockroach connections
	Redis.Init(running)     //redis connections
	initCassandraSession()  //cassandra
	Couch.Init()            //couchbase
}

/*! \brief Once we're done, this closes the open connections to things
 */
func CloseConnections() {
	cockroach.Close()
	//cassandraCloseDown()
	Redis.disconnect()
	//Couch.CloseDown()
}

/*! \brief Checks to see if the needle is locacted in the haystack slice
 */
func InArray(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false //couldn't find it
}

/*! \brief Checks to see if the needle is locacted in the haystack slice
 */
func InArray64(haystack []int64, needle int64) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false //couldn't find it
}

/*! \brief Removes the duplicates in an array
 */
func removeDuplicates(a []string) string {
	result := []string{}
	seen := map[string]string{}
	for _, val := range a {
		if len(val) > 0 {
			if _, ok := seen[val]; !ok {
				result = append(result, val)
				seen[val] = val
			}
		}
	}
	return strings.Join(result, ",")
}

/*! \brief Apparently splittin an empty string gives an array of 1 back
 */
func Split(str string) []string {
	ret := make([]string, 0)
	str = strings.Trim(str, " ,")
	if len(str) > 0 {
		ret = strings.Split(str, ",")
	}
	return ret
}

func nullable (in string) (out sql.NullString) {
	if len(in) > 0 {
		out.String = in
		out.Valid = true
	}
	return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- reused FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Preps a string for use in the database by removing "weird" characters
 */
func safeRegex(in string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9]")
	return strings.ToLower(r.ReplaceAllString(in, ""))
}

func rowsChk(rows *sql.Rows) error {
	rows.Close() //close it here, no harm in doing it more than once
	return ErrChk(rows.Err())
}

func emptyChk(err error) error {
	if err != nil {
		if err != sql.ErrNoRows { //not an error we can ignore
			return ErrChk(err)
		}
	}
	return nil //we're done, no error
}

func dateOffset (days int) string {
    return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func prevDate (date string) string {
    tm, err := time.Parse("2006-01-02", date)
    if err == nil {
        old := tm.AddDate(0, 0, -1)
        return old.Format("2006-01-02")
    } else {
        return dateOffset (-1)
    }
}

func validateDateRange (start, end *string) (time.Time, time.Time, bool) {
    startTm, err := time.Parse("2006-01-02", *start)
    if err != nil { startTm, _ = time.Parse("2006-01-02", dateOffset(1)) }

    endTm, err := time.Parse("2006-01-02", *end)
    if err != nil { endTm, _ = time.Parse("2006-01-02", dateOffset(-30)) }

    if endTm.After(startTm) {
        *start = startTm.Format("2006-01-02")
        *end = endTm.Format("2006-01-02")
        return startTm, endTm, true
    } else {
        *end = startTm.Format("2006-01-02")
        *start = endTm.Format("2006-01-02")
        return startTm, endTm, false
    }
}

func minuteOffset (minute int) string {
    return time.Now().Local().Add(time.Minute * time.Duration(minute)).Format("2006-01-02 15:04:05")
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- WRAPPER FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func Err(msg string, params ...interface{}) error {
	return toolz.Err(msg, params...)
}

func UM(jAttr []byte, out interface{}) error {
	return toolz.UM(jAttr, out)
}

func ErrChk (err error) error {
	return toolz.ErrChk(err)
}

func StrToInt64(in string) int64 {
	return toolz.StrToInt64 (in)
}
