/*! \file user.go
 *  \brief All things related to the user api calls
 *  This is the main entry point for /user/ urls
 *
 */

package web

import (
    //"fmt"
	"encoding/json"
	
	"github.com/NathanRThomas/boiler_api/db"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! Main class for the user file
 */
type User_c struct {
	User        *db.User_t
	Post        []byte
    users       db.User_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Favorites the target user.  So that this user is now following them
 */
func (this *User_c) update () ([]byte, error) {
    var m struct { First, Last string }
	UM(this.Post, &m)	//parse our object

    
    return json.Marshal(APIResponse_t{false, "", api_no_error})
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point for the user section of the api.  Any url that starts with "user" goes here
 */
func (this *User_c) Entry(statement string) ([]byte, error) {
    
    switch statement {
    case "update":
        return this.update()
    }

    return json.Marshal(APIResponse_t{false, "Unknown call", api_error_unknown_call})
}