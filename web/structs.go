/*! \file structs.go
  \brief Re-used structs and defines for the api
*/

package web

import (
	
	"github.com/NathanRThomas/boiler_api/toolz"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

//! API response error codes
const (
	api_no_error = iota //everythign is good, this has a value of 0
	api_error_unknown_call
	API_error_bad_call
	API_error_bad_marshal
	Api_error_not_logged_in
	api_error_missing_param
    api_authy                   // 6 means i need the user to authorize this request
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- API OBJECTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

//! Basic response object that we send back when there's nothing else to be said
type APIResponse_t struct {
	Success   bool   `json:"success"`
	Msg       string `json:"msg,omitempty"`
	ErrorCode int    `json:"error"`
}

type PublicEntry interface {
	Entry(string) ([]byte, error)
}

func UM(jAttr []byte, out interface{}) error {
	return toolz.UM(jAttr, out)
}