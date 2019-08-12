/*! \file db.go
    \brief Special couchbase database stuff

https://godoc.org/gopkg.in/couchbase/gocb.v1
https://docs.couchbase.com/go-sdk/1.5/start-using-sdk.html
https://docs.couchbase.com/server/6.0/getting-started/do-a-quick-install.html

*/

package db

import (
    "fmt"
    "time"
    "log"
	
    "github.com/NathanRThomas/boiler_api/toolz"
    
	"gopkg.in/couchbase/gocb.v1"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const couchCacheTime = 60   // time in seconds for our default cache to expire
const couchListSize = 99999

type couchbase_c struct { 
    cache *gocb.Bucket
    cluster *gocb.Cluster
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- INIT FUNCTIONS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *couchbase_c) Init () {
    var err error
    if this.cluster == nil && len(toolz.AppConfig.Couch.IP) > 0 {
        this.cluster, err = gocb.Connect("couchbase://" + toolz.AppConfig.Couch.IP)
        if err == nil {
            this.cluster.Authenticate(gocb.PasswordAuthenticator{
                Username: toolz.AppConfig.Couch.Username,
                Password: toolz.AppConfig.Couch.Password,
            })

            this.cluster.SetN1qlTimeout (5 * time.Second)    //set a timeout
            this.cluster.SetServerConnectTimeout (2 * time.Second)   

            this.cache, err = this.cluster.OpenBucket("cache", "")
            //this.cache.Manager("", "").CreatePrimaryIndex("", true, false)
        }
    }
    if err != nil {
        log.Panic("Couchbase failed ", err)
    }
    return //we're done
}

/*! \brief Closes down the connection pool when we shutdown the service
*/
func (this *couchbase_c) CloseDown () {
    if this.cluster != nil { 
        this.cluster.Close() 
        this.cluster = nil //we're done
        this.cache = nil
    }
}

//----- Caching ------------------------------------------------------------------------------------------------------//
func (cb *couchbase_c) SetCache (key string, val interface{}, exp uint32) error {
	if exp == 0 { exp = couchCacheTime }
	_, err := cb.cache.Upsert(key, val, exp)
    return ErrChk(err)
}

func (cb *couchbase_c) GetCache (key string, val interface{}) bool {
    cas, _ := cb.cache.Get (key, val)
    return cas != 0
}

/*! \brief I seem to be doing this a lot so i created this
	This will check to see if a value exists, and if so return true
	if it isn't set, then it will set it and return false
*/
func (cb *couchbase_c) Flagged (key string, exp uint32) bool {
	val := 0
	cb.GetCache (key, &val)
	if val > 0 { return true }	//it exists already
	
	if exp == 0 { return false }	//if this is zero, then we dont' want to set it, we just wanted to see if it was set
	
	val = 1
	cb.SetCache (key, val, exp)
	return false 	//didn't exist yet
}

func (cb *couchbase_c) ClearKey (msg string, params ...interface{}) {
	cb.cache.Remove (fmt.Sprintf(msg, params...), 0)
}

func (cb *couchbase_c) ListPrepend (key string, val interface{}, maxSize uint) error {
	if maxSize == 0 { maxSize = couchListSize }
	_, err := cb.cache.ListPrepend(key, val, true)
	var sz uint
	if err == nil {
		sz, _, err = cb.cache.ListSize (key)
		if err == nil {
			if sz > maxSize {
				_, err = cb.cache.ListRemove (key, sz-1)	//remove the last index
			}
		}
	}
    return ErrChk(err)
}

var Couch couchbase_c

func TestCouchbase () (err error) {
	if err = Couch.SetCache("testing", "1", 10); err == nil {
		val := ""
		if !Couch.GetCache ("testing", &val) {
			err = fmt.Errorf("cache not found")
		}
	}
	return
}