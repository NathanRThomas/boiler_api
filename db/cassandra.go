/*! \file db.go
    \brief Special cassandra database stuff

*/

package db

import (
    "fmt"
	"strings"
    "time"
    "log"
	
    "github.com/NathanRThomas/boiler_api/toolz"
    
	"github.com/gocql/gocql"
)

var session *gocql.Session

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- INIT FUNCTIONS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func initCassandraSession () (err error) {
    if session == nil || session.Closed() {
		if len(toolz.AppConfig.Cassandra.Ips) > 0 {
			cluster := gocql.NewCluster(toolz.AppConfig.Cassandra.Ips...)
			cluster.Keyspace = toolz.AppConfig.Cassandra.Keyspace
			cluster.Consistency = gocql.Quorum
			//cluster.NumConns = 1
			cluster.Timeout = time.Second * 15
			session, err = cluster.CreateSession()
		}
    }
    if err != nil {
        log.Panic("Cassandra failed ", err)
    }
    return
}

/*! \brief Closes down the connection pool when we shutdown the service
*/
func cassandraCloseDown () {
	if session != nil && !session.Closed() { session.Close() }
}

func cassChk (err error) error {
    if err != nil {
        switch e := err.(type) {
        case *gocql.RequestErrUnavailable:
            return ErrChk(fmt.Errorf(e.String()))
        case *gocql.RequestErrWriteTimeout:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrWriteFailure:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrReadTimeout:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrAlreadyExists:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrUnprepared:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrReadFailure:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        case *gocql.RequestErrFunctionFailure:
            return ErrChk(fmt.Errorf("%s :: %s", err.Error(), e.String()))
        default:
            if err.Error() == "not found" { //ignore not found errors
                return nil
            } else {
                return ErrChk(fmt.Errorf("Unknown cassandra error :: %s", err.Error()))
            }
        }
    } else {
        return nil
    }
}

func cassExec (query string, args ...interface{}) error {
    return cassChk(session.Query(query, args...).Exec())
}

func cassSlice (query string, args ...interface{}) ([]map[string]interface{}) {
    slice, err := session.Query(query, args...).Iter().SliceMap()
    ErrChk(err)
    return slice
}

func cassDateMS (date string) (out string) {
    if len(date) > 23 {
        out = date[:23]
    }
    return strings.Replace(out, "T", " ", -1)
}

func cassTimeUUID (timeStr string) gocql.UUID {
    if len(timeStr) > 0 {
        tm, err := time.Parse("2006-01-02 15:04:05", timeStr)
        if err == nil {
            return gocql.UUIDFromTime(tm)
        } else {
            ErrChk(err)
        }
    }
    return gocql.TimeUUID()
}
