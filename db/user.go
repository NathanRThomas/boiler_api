/*! \file user.go
 *  \brief For the DB package, this contains calls specific for the user package
 *  Just a way to try to keep the file size managable
 */

package db

import (
	//"fmt"
	//"encoding/json"
	
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const (
    userStatus_deleted         = 1 << iota     
    
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC STRUCTS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type User_t struct {
	ID		string 	`json:"user_id"`	//i pulled these out cause they go into the top level of the api calls
	Token	string 	`json:"user_token"` 
	Email string
	Attr struct {
		First, Last 	string
	}
	Mask 	int64
}

type User_c struct {
	cache 	cache_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Returns info about the user based on the id
 *	cached, use for normal user api calls
 */
func (this *User_c) Get (user *User_t) (ok bool) {
    if len(user.ID) == 0 { return }  //this isn't good, can't find a user with the id this way
    
    //now, let's see if we're cached already, we need to do it now cause it's based on the device id and type
    key := this.cache.UserKey (user.ID)
    if Redis.GetCache(key, user) { //check for a cache hit
        return true //we're good!
	}
	
	var jAttr []byte
	err := sqlDB.QueryRow(`SELECT email, attributes, mask, token FROM users WHERE id = $1`, user.ID).Scan(&user.Email, &jAttr, &user.Mask)

	if err == nil {
		Redis.SetWithTimeout(0, key, *user, 60) //cache this result for next time
		ok = true   //we're good
	}
    return
}

/*! \brief Updates the info about this user as well as their associated subscriber
*/
func (this *User_c) Update (user *User_t) (error) {
    defer this.cache.ClearUser(user.ID)
    return nil
}
