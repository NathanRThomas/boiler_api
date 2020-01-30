/*! \file handlers.go
	\brief Contains all handlers for incoming http requests

	Created 2019-10-24 by NateDogg
*/

package main

import (
	//"github.com/NathanRThomas/boiler_api/cmd"
	//"github.com/NathanRThomas/boiler_api/pkg/models"

	//"github.com/pkg/errors"
	"github.com/dgrijalva/jwt-go"
			
	//"fmt"
	"net/http"
	"encoding/json"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFAULT -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app_c) defaultHandler (w http.ResponseWriter, r *http.Request) {
	this.NotFound (w)
}


func (this *app_c) loggedIn (w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you're logged in"))
}

func (this *app_c) login (w http.ResponseWriter, r *http.Request) {
	token, err := this.CreateJWT ("123")
	if err != nil {
		this.ServerError (err, w)
		return
	}

	jAttr, _ := json.Marshal(token)
	w.Write(jAttr)
}

func (this *app_c) info (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jwt := ctx.Value("jwt").(*jwt.Token).Claims.(jwt.MapClaims)

	jAttr, _ := json.Marshal("you're in! " + jwt["id"].(string))
	w.Write(jAttr)
}