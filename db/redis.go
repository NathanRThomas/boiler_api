/*! \file redis.go
  \brief Handles any redis connection stuff and calls to it
*/

package db

import (
	"fmt"
	"log"
	"strconv"

	"github.com/NathanRThomas/boiler_api/toolz"
	"github.com/mediocregopher/radix.v2/pool"
)

const maxRedisPoolSize = 10 //max number of cache threads waiting in the pool
const defaultCacheTime = 20   //time in seconds for our default cache to expire
const longCacheTime = 2419200 //4 weeks -- max = 2591999 30 days - 1 second

type redisIndex int

/* NOTE!!
 * This order must match the order of the "redis" array in the config file
 *
 */
const (
	redis_index_cache        redisIndex = iota //6379
	redis_index_tasks                          //6380
)

type redis_c struct {
	pools     	[]*pool.Pool //list of our active pools
	shrinkage 	toolz.Shrinkage_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE CACHE FUNCTIONS -------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- SPECIFIC FUNCTIONS ------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this redis_c) Zrange(idx redisIndex, key string) []string {
	rs, _ := this.pools[idx].Cmd("ZRANGE", key, 0, -1).List()
	//fmt.Println(js)
	//l, _ := rs.List()
	return rs
	/*
		l, _ := rs.List()
		for _, elemStr := range l {
			fmt.Println(elemStr)
		}
		return nil
	*/
}

func (this redis_c) Zadd(idx redisIndex, key, val string, score int64) {
 this.pools[idx].Cmd("ZADD", key, score, val)
}

func (this redis_c) Zrem(idx redisIndex, key, val string) {
 this.pools[idx].Cmd("ZREM", key, val)
}

func (this redis_c) Zrank(idx redisIndex, key, val string) int {
	rs, err := this.pools[idx].Cmd("ZRANK", key, val).Int()
	if err != nil {
		return -1
	} //not ranked
	return rs //else just return the rank
}

func (this redis_c) ZrangeByScore(idx redisIndex, key string, max int) []string {
	rs, _ := this.pools[idx].Cmd("ZRANGEBYSCORE", key, "(", max).List()
	return rs
}

func (this redis_c) ZRemoveByRank(idx redisIndex, key string, limit int) {
 this.pools[idx].Cmd("ZREMRANGEBYRANK", key, limit, limit*2)
}

func (this redis_c) ZRemoveByScore(idx redisIndex, key string, min, max int64) {
 this.pools[idx].Cmd("ZREMRANGEBYSCORE", key, min, max)
}

func (this redis_c) Lpush(idx redisIndex, key, val string) {
 this.pools[idx].Cmd("LPUSH", key, val)
}

func (this redis_c) Rpush(idx redisIndex, key, val string) {
 this.pools[idx].Cmd("RPUSH", key, val)
}

func (this redis_c) Llen(idx redisIndex, key string) int {
	rs, _ := this.pools[idx].Cmd("LLEN", key).Int()
	return rs
}

func (this redis_c) Sadd(idx redisIndex, key, val string) {
 this.pools[idx].Cmd("SADD", key, val)
}

func (this redis_c) Spop(idx redisIndex, key string) []byte {
	rs, _ := this.pools[idx].Cmd("SPOP", key).Bytes()
	return rs
}

func (this redis_c) Rpop(idx redisIndex, key string) []byte {
	rs, _ := this.pools[idx].Cmd("RPOP", key).Bytes()
	return rs
}

func (this redis_c) Get(idx redisIndex, key string) []byte {
	rs, _ := this.pools[idx].Cmd("GET", key).Bytes()
	return rs
}

func (this redis_c) Set(idx redisIndex, key, val string) error {
	return this.pools[idx].Cmd("SET", key, val).Err
}

func (this redis_c) Expire(idx redisIndex, key string, timeout int) {
 this.pools[idx].Cmd("EXPIRE", key, timeout)
}

func (this redis_c) Del(idx redisIndex, key string) {
 this.pools[idx].Cmd("DEL", key)
}

func (this redis_c) SetEx(idx redisIndex, key string, timeout int, val string) {
 this.pools[idx].Cmd("SETEX", key, timeout, val)
}

func (this redis_c) Zincr(idx redisIndex, key, val string, weight int) {
 this.pools[idx].Cmd("ZINCRBY", key, weight, val)
}

func (this redis_c) Incr(idx redisIndex, key string) int {
	rs, _ := this.pools[idx].Cmd("INCR", key).Int()
	return rs
}

func (this redis_c) Incrby(idx redisIndex, key, value string) {
 this.pools[idx].Cmd("INCRBY", key, value)
}

func (this redis_c) KeySearch(idx redisIndex, key string) []string {
	rs, _ := this.pools[idx].Cmd("KEYS", key+"*").List()
	return rs
}

func (this redis_c) Ping(idx redisIndex) string {
	rs, _ := this.pools[idx].Cmd("PING").Bytes()
	return string(rs[:])
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this redis_c) DDosCheck(ip, email string) bool {
	key := fmt.Sprintf("ddos:%s", ip)
	emailKey := fmt.Sprintf("ddos:%s", email)
 	this.Incr(redis_index_cache, key)                                          //incr this user
 	this.Incr(redis_index_cache, emailKey)                                     //incr this user
	cnt, _ := strconv.Atoi(string(this.Get(redis_index_cache, key)))           //now get it out
	emailCnt, _ := strconv.Atoi(string(this.Get(redis_index_cache, emailKey))) //now get it out
	if cnt > 16 {
		cnt = 16
	}
	if emailCnt > 16 {
		emailCnt = 16
	} //no reason to go crazy here

 this.Expire(redis_index_cache, key, (1<<uint(cnt/2))+2)           //keep a scrolling window
 this.Expire(redis_index_cache, emailKey, (1<<uint(emailCnt/2))+2) //keep a scrolling window

	if cnt < 5 && emailCnt < 5 {
		return true
	}
	return false //locked out
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CACHE FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Our default function, this sets a value associated with a key for our default duration time
 *  NOTE this also sets things to expire with a timeout, it's the default timeout set above
 */
func (this redis_c) SetCache(key string, val interface{}) (err error) {
	return this.SetWithTimeout(redis_index_cache, key, val, defaultCacheTime)
}

/*! \brief Sets a value to a key for a specified period of time.  When we're setting things where the default timeout doesn't apply
 */
func (this redis_c) SetWithTimeout(poolIdx redisIndex, key string, val interface{}, timeout int) (err error) {
	if len(key) == 0 {
		return
	} //just a double check

	resp := this.pools[poolIdx].Cmd("SETEX", key, timeout, this.shrinkage.Compress(val))
	if resp.Err != nil {
		err = fmt.Errorf("db.Cache SetWithTimeout Setex :: %s", resp.Err.Error())
	}
	return //and done
}

/*! \brief Doesn't make any sense to set things without the ability to get them as well
 */
func (this redis_c) GetCache(key string, val interface{}) bool {
	return this.GetCacheFromIndex(redis_index_cache, key, val)
}

func (this redis_c) GetCacheFromIndex(poolIdx redisIndex, key string, val interface{}) bool {
	if len(key) == 0 {
		return false
	} //just a double check

	js, err := this.pools[poolIdx].Cmd("GET", key).Bytes()
	if err != nil {
		return false
	} //this means that we had a cache miss
 this.shrinkage.Uncompress(js, val)
	return true //we're good
}

func (this redis_c) ClearKey(msg string, params ...interface{}) {
 this.clearKeyFromIndex(redis_index_cache, msg, params...)
}

func (this redis_c) clearKeyFromIndex(poolIdx redisIndex, msg string, params ...interface{}) {
	if len(msg) > 0 {
	 this.pools[poolIdx].Cmd("DEL", fmt.Errorf(msg, params...))
	}
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- INIT FUNCTIONS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *redis_c) disconnect() {
	for _, p := range this.pools {
		p.Empty()
	}
}

func (this *redis_c) connect() (err error) {
	for idx, port := range toolz.AppConfig.Redis.Ports { //loop through the ports
		if pPool, err := pool.New("tcp", fmt.Sprintf("%s:%d", toolz.AppConfig.Redis.IP, port), maxRedisPoolSize); err == nil {
		 	this.pools[idx] = pPool //set our new connected pool
		} else {
			return Err("redisSetConnList: " + toolz.AppConfig.Redis.IP + " :: " + err.Error()) //this is bad
		}
	}
	return nil //all ports were able to be opened
}

/*! \brief Sets up all the connection pools that we'll need in the future
 */
func (this *redis_c) Init(running *bool) {
	if len(toolz.AppConfig.Redis.IP) == 0 { return } //can't do anything
	if len(toolz.AppConfig.Redis.Ports) == 0 { log.Panic("There are no ports listed in the config for redis") } //can't do anything

 	this.pools = make([]*pool.Pool, len(toolz.AppConfig.Redis.Ports)) //init this based on the number of ports/index we're going to have

	if err := this.connect(); err == nil { //try to connect via our active index
		return                  //we're good
	}
	//if we're here it's cause there were no indexes that worked, so we have to bail
	log.Panic("All redis connections failed")
}

var Redis redis_c //init our class for handling redis
