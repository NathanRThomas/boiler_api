/*! \file routes.go
	\brief Pulls out the routing of the urls to functions
*/

package main

import (
	//"fmt"
	"net/http"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app_c) notFound (w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Define our routes for this service
	this one is pretty simple, just a 404 for any other endpoints
*/
func (this *app_c) routes () http.Handler {
	mux := this.Routes () // get our base mux for handling things

	mux.HandleFunc("/", this.notFound)	// just return 404
	
    return mux
}
