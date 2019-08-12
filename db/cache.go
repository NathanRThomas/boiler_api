/*! \file cache.go
 *  \brief To keep all those cache calls in one place at least
 	Switch this between couch and redis depending on your application
 */

package db

import (
	"fmt"
    
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC STRUCTS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type cache_c struct {
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


//-- Users ---------------------------------------------------------------------------------------------------------------//
func (this *cache_c) UserKey (userID string) string {
	return fmt.Sprintf("uid:%s", userID)
}

func (this *cache_c) ClearUser (userID string) {
    Redis.ClearKey(this.UserKey(userID))
}


//-- Login ---------------------------------------------------------------------------------------------------------------//
func (this *cache_c) LoginKey (userID string) string {
	return fmt.Sprintf("u:%s", userID)
}

func (this *cache_c) ClearLogin (userID string) {
	Redis.ClearKey (this.LoginKey (userID))
	this.ClearUser (userID)
}