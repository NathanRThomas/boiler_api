/*! \file user.go
	\brief Handlers for user endpoints
*/

package main

import (
	"github.com/NathanRThomas/boiler_api/cmd"
	"github.com/NathanRThomas/boiler_api/pkg/models"
	
	"github.com/pkg/errors"
			
	//"fmt"
	"net/http"
	"database/sql"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- Not Logged In -----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Attempts to log in a user based on what they've passed us
*/
func (this *app_c) userLogin (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := &models.User_t{}
	err := this.ParseFromBody (ctx, user)
	if err != nil {	this.ErrorWithMsg (err, w, http.StatusBadRequest, cmd.ApiErrorCode_parsingRequestBody, ""); return } // bail here

	err = this.Users.Login (user)

	switch errors.Cause (err) {
	case nil: // it worked

	case sql.ErrNoRows: // no user found
		err = errors.Wrap (models.ErrType_returnToUser, "Info not found in our system") 

	default: // just pass this error through
	}

	this.Respond (err, w, struct { // either it worked or it didn't, pass it out
		User   *models.User_t
	} { user })
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- Logged In ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Just returns the info about our target user
*/
func (this *app_c) userGet (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	user, ok := ctx.Value("user").(*models.User_t) // get our current user
	if !ok { this.ServerError (errors.WithStack (models.ErrType_userMissing), cmd.ApiErrorCode_missingFromContext, w); return } 
	
	this.Respond (nil, w, user) // we're done
}
