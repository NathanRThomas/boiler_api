/*! \file main.go
  \brief Handles any redis connection stuff and calls to it
*/

package redis

import (

	"github.com/mediocregopher/radix/v3"
	"github.com/pkg/errors"

	"fmt"
	"log"
	"encoding/json"
	"strings"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var ErrServiceDown = errors.New("service is down")
var ErrNoServiceAvailable = errors.New("no service is connected")
var ErrKeyNotFound = errors.New("key not found or has expired")
var ErrPingFailed = errors.New("Ping did not Pong")

const MaxPoolSize = 10     //max number of cache threads waiting in the pool
const defaultCacheTime = 30 	//time in seconds for our default cache to expire

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type redisQue_t struct {
	Key, Val string
}

type DB_c struct {
	DB 			*radix.Pool
    pushes 		[]redisQue_t	//list of items that didn't get queued
}

/*! \brief We "expect" there to be a connection refused when we have our redis server down, otherwise we want to handle errors from redis
*/
func (this *DB_c) locErr (err error) error {
	if err != nil { 
		if strings.Contains (err.Error(), "connection refused") || strings.Contains (err.Error(), "client is closed") {
			err = ErrServiceDown
		}
	}

	return err
}

func (this *DB_c) getHandleNil (command, key string, val interface{}) error {
	if this.DB == nil { return errors.WithStack (ErrNoServiceAvailable) }
	loc := ""
	mn := radix.MaybeNil{Rcv: &loc}
	if err := this.DB.Do(radix.Cmd(&mn, command, key)); err == nil {
		if mn.Nil || len(loc) == 0 { return errors.WithStack (ErrKeyNotFound) }
		return errors.Wrapf(json.Unmarshal([]byte(loc), val), " %s : %s ", key, loc)	//wrap this with whatever the string we failed to extract was
	} else {
		return errors.WithStack(this.locErr(err))	// see if this is worth recording
	}
}

func (this *DB_c) js (in interface{}) string {
	jAttr, err := json.Marshal (in)
	if err != nil { 
		log.Printf ("%s\n%+v\n", err.Error(), in) 
		return ""
	}
	return string(jAttr)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- SPECIFIC FUNCTIONS ------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *DB_c) lpush (key string, val interface{}) (err error) {
	if this.DB != nil {
		err = errors.WithStack(this.locErr(this.DB.Do(radix.FlatCmd(nil, "LPUSH", key, this.js(val))))) // either it was good or we can't handle the error
		if err == nil { return }	//it worked
	} else {
		err = errors.WithStack (ErrServiceDown)
	}

	//in these cases we can save this and try to insert it later
	this.pushes = append (this.pushes, redisQue_t { Key: key, Val: this.js(val) })
	return 
}

func (this *DB_c) rpop (key string, val interface{}) error {
	return this.getHandleNil ("RPOP", key, val)
}

func (this *DB_c) get (key string, val interface{}) error {
	return this.getHandleNil ("GET", key, val)
}

func (this *DB_c) set (key string, timeout int, val interface{}) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.FlatCmd(nil, "SETEX", key, fmt.Sprintf("%d", timeout), this.js(val)))) == nil
}

func (this *DB_c) del (key string) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.FlatCmd(nil, "DEL", key))) == nil
}

func (this *DB_c) llen (key string) (size int) {
	if this.DB == nil { return }
	this.locErr(this.DB.Do(radix.Cmd(&size, "LLEN", key)))
	return
}

func (this *DB_c) sadd (key string, val interface{}) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.FlatCmd(nil, "SADD", key, this.js(val)))) == nil
}

func (this *DB_c) spop (key string, val interface{}) error {
	return this.getHandleNil ("SPOP", key, val)
}

func (this *DB_c) zadd (key string, score int64, val string) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.Cmd(nil, "ZADD", key, fmt.Sprintf("%d", score), val))) == nil
}

func (this *DB_c) zrange (key string, score int64) []string {
	ret := make([]string, 0)
	this.locErr(this.DB.Do(radix.Cmd(&ret, "ZRANGE", key, "0", fmt.Sprintf("%d", score))))
	return ret
}

func (this *DB_c) zrem (key, member string) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.FlatCmd(nil, "ZREM", key, member))) == nil
}

func (this *DB_c) expire (key string, timeout int) bool {
	if this.DB == nil { return false }
	return this.locErr(this.DB.Do(radix.FlatCmd(nil, "EXPIRE", key, fmt.Sprintf("%d", timeout)))) == nil
}

func (this *DB_c) incr (key string) (out int) {
	if this.DB == nil { return }
	this.locErr(this.DB.Do(radix.Cmd(&out, "INCR", key)))
	return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TESTING -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *DB_c) Ping () error {
	if this.DB == nil { return nil } // this is bad

	out := ""
	if err := this.locErr(this.DB.Do(radix.Cmd(&out, "PING"))); err == nil {
		if out == "PONG" { return nil }
		return errors.WithStack (ErrPingFailed)
	} else {
		return err
	}
}

/*	
	if goodCnt == len(toolz.AppConfig.Redis.IPs) {	//we're all good, add in our "pending" channels
		for len(this.pushes) > 0 {	// we recovered from an error, que the remaining items
			push := <-this.pushes // get the next one
			err := this.pools[this.pushIdx()].Do(radix.FlatCmd(nil, "LPUSH", push.Key, push.Val))
			if err != nil { ErrChk(err); return }	//not sure, just bail
		}
	}
*/
