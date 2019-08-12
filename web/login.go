 /*! \file login.go
 *  \brief web calls related to broadcasts
 *
 */

package web

import (
	"encoding/json"
	//"fmt"
	
	"github.com/NathanRThomas/boiler_api/db"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! Main class for the user file
 */
type Login_c struct {
	Post	[]byte
	db 		db.Login_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *Login_c) login () ([]byte, error) {
    var m struct { Email, Password string }
	UM(this.Post, &m)  //parse our object
	
	if user, ok := this.db.Email (m.Email, m.Password); ok {
		return json.Marshal(struct {
			APIResponse_t
			User     db.User_t
		}{APIResponse_t{true, "Logged in successfully!", api_no_error}, *user})
	} else {
        return json.Marshal(APIResponse_t{false, "Email or Password is invalid", api_no_error})
    }
}

func (this *Login_c) create () ([]byte, error) {
    var m struct { Email, Password string }
	UM(this.Post, &m)  //parse our object
	
	if user, err := this.db.Create (m.Email, m.Password); err == nil {
		return json.Marshal(struct {
			APIResponse_t
			User     db.User_t
		}{APIResponse_t{true, "Account created successfully!", api_no_error}, *user})
	} else {
		return json.Marshal(APIResponse_t{false, err.Error(), api_no_error})	//pass this error back out
	}
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point for the user section of the api.  Any url that starts with "user" goes here
 */
func (this *Login_c) Entry(statement string) ([]byte, error) {
    
	switch statement {
	case "login":
		return this.login()
    case "create":
        return this.create()
    }
	return json.Marshal(APIResponse_t{false, "Unknown call", api_error_unknown_call})
}