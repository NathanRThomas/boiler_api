/*! \file routes.go
	\brief Pulls out the routing of the urls to functions
*/

package main

import (
	"github.com/NathanRThomas/boiler_api/cmd"

	"github.com/urfave/negroni"

	//"fmt"
	"net/http"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type handler_c struct {
	cmd.Handler_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app_c) routes () http.Handler {
	handler := &handler_c { 
		Handler_c: this.NewHandler(),
	}
	mux := handler.Routes () // get our base mux for handling things

// Default handler
	mux.HandleFunc("/", this.defaultHandler)	// just return 404

// logged in
	std := handler.ApiChain ()	// standard chain that all calls make

	mux.Handle ("/login", std.ThenFunc(this.login)).Methods (http.MethodPost)

	jwtMiddleware := this.GetJWTMiddleware ()

	mux.Handle("/info", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(std.ThenFunc(this.info)),
	)).Methods (http.MethodGet)
	
	return mux
}

