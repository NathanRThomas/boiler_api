/*! \file cache.go
  \brief cache specifc redis calls
*/

package redis

import (
	"github.com/pkg/errors"

	"fmt"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CACHE -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Sets the cache
*/
func (this *DB_c) SetCache (key string, val interface{}, timeout int) bool {
	if timeout <= 0 { timeout = defaultCacheTime }
	return this.set (key, timeout, val)
}

/*! \brief Doesn't make any sense to set things without the ability to get them as well
*/
func (this *DB_c) GetCache (key string, val interface{}) error {
	return this.get (key, val)
}

func (this *DB_c) ClearKey (msg string, params ...interface{}) {
    this.del(fmt.Sprintf(msg, params...))
}

/*! \brief Handles checking to see if the key exists, and if not sets it
	returns true if it's already set, false if it's not set yet
*/
func (this *DB_c) Flagged (key string, exp int) bool {
	val := ""
	if errors.Cause(this.GetCache (key, &val)) == ErrKeyNotFound {
		if exp == 0 { exp = 3600 }	//default to 1 hour
		//doesn't exist, so set it for next time
		this.SetCache (key, "1", exp)
		return false
	}
	return true //already set
}
