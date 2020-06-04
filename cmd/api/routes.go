/*! \file routes.go
	\brief Pulls out the routing of the urls to functions
*/

package main

import (
	
	//"fmt"
	"net/http"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app_c) routes () http.Handler {
	mux := this.Routes () // get our base mux for handling things

	// Default handler
	std := this.ApiChain ()	// standard chain that all calls make
	ddos := std.Append (this.Ddos)

	loggedIn := std.Append (this.bearerCheck)	// validates the bearer token


// user - not logged in
	mux.Handle("/login", ddos.ThenFunc (this.userLogin)).Methods(http.MethodPut, http.MethodOptions)

// user - logged in
	mux.Handle("/user", loggedIn.ThenFunc (this.userGet)).Methods(http.MethodGet, http.MethodOptions)

	return mux
}

