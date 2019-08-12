/*! \file login.go
 *  \brief For the DB package
 */

package db

import (
	"fmt"
    "strings"
    
	"github.com/NathanRThomas/boiler_api/toolz"
)


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC STRUCTS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Login_c struct {
	mailman 	toolz.Mailman_c
	cache 		cache_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Generates our password hash with our salt
*/
func (this *Login_c) passHash (password string) string {
	return toolz.Hash(password + salt, 0)
}

func (this *Login_c) tokenHash (password, email string) string {
	return toolz.Hash(password + email + salt, 0)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Tries to log a user in with an email/password and returns the admin object
    This caches the result if it's successful
*/
func (this *Login_c) Email (email, password string) (user *User_t, exists bool) {
    if len(password) > 0 && this.mailman.ValidateEmail (&email) {
		user = &User_t{}	//init some memory here
        emptyChk(sqlDB.QueryRow(`SELECT id, token FROM users WHERE password = $2 AND email = $1 AND mask & $3 = 0`,
                            email, this.passHash(password), userStatus_deleted).Scan(&user.ID, &user.Token))
		
		if len(user.ID) > 0 { exists = this.Token (user) }	//get the rest of the info about this user
    }
    return
}

/*! \brief handles logging into the site, and then caching that result
        This is cached without the org, as it's a pointer, and the target bot id needs to be tracked properly
 */
func (this *Login_c) Token (user *User_t) (bool) {
    //make sure we have good data
    if len(user.ID) > 0 && len(user.Token) > 0 {
        //first, let's see if we're cached already
        key := this.cache.LoginKey (user.ID)
		if Redis.GetCache(key, user) { //check for a cache hit
			return true 	//we're good
		}
		
		//try the database now
		var jAttr []byte
		err := sqlDB.QueryRow(`SELECT email, attributes, mask FROM users WHERE id = $1 AND token = $2 AND mask & $3 = 0`,
							user.ID, user.Token, userStatus_deleted).Scan(&user.Email, &jAttr, &user.Mask)

		if err == nil { //we were able to login, now cache this for next time
			UM(jAttr, &user.Attr)
			Redis.SetWithTimeout(0, key, user, 60) //cache this result for next time
			return true
		}
	}
    return false //couldn't login
}

/*! \brief Resets a user's password
 */
func (this *Login_c) ResetPassword (password string, user *User_t) (err error) {
    //try to login with this
    if this.Token(user) {
        if err = toolz.ValidatePassword (password); err == nil {
            //copy this new stuff back out
            user.Token = this.tokenHash(password, user.Email)
            err = dbExec(`UPDATE users SET password = $1, token = $2 WHERE id = $3`,
							this.passHash(password), user.Token, user.ID)
            if err == nil {
                this.cache.ClearLogin (user.ID)
                return nil
            }
		}
    } else {
        err = fmt.Errorf("Invalid user information.  Please try the email again")
	}
	return
}

/*! \brief Checks to see if this email is unique in the system
	returns true if it's unique
*/
func (this *Login_c) UniqueEmail (email, id string) bool {
	exists := ""
    emptyChk(sqlDB.QueryRow(`SELECT id::text FROM users WHERE email = lower($1) AND status_mask & $3 = 0 AND id <> $2`, 
                      strings.Trim(email, " "), id, userStatus_deleted).Scan(&exists))
    return len(exists) == 0
}

/*! \brief Creates a new user
*/
func (this *Login_c) Create (email, password string) (*User_t, error) {
	err := toolz.ValidatePassword(password)
	if err != nil { return nil, err }	//bad password

	if this.mailman.ValidateEmail (&email) == false { return nil, fmt.Errorf("Email appears invalid") }	//bad email

	if this.UniqueEmail (email, "") == false { return nil, fmt.Errorf("Email already in use by another user") }	//not unique email
	
	user := &User_t { Token : this.tokenHash(password, email) }	//generate our object
		
	err = toolz.ErrChk(sqlDB.QueryRow(`INSERT INTO users (email, password, token) VALUES ($1, $2, $3) RETURNING id`,
						email, this.passHash(password), user.Token).Scan(&user.ID))
	if err == nil && len(user.ID) > 0 {	//inserted successfully
		if this.Token (user) {
			return user, nil 	//we're good
		}
	}
	return nil, fmt.Errorf("Database error occured")	//not sure what to tell the user on this one
}

