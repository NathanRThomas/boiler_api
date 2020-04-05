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

	input := &cmd.SignupUser_t{}
	err := this.ParseFromBody (ctx, input)
	if err != nil {	this.ErrorWithMsg (err, w, http.StatusBadRequest, cmd.ApiErrorCode_parsingRequestBody, ""); return } // bail here

	user, err := this.Users.Login (input.Email, input.Password)
	switch errors.Cause (err) {
	case models.ErrType_noIdentifiers, sql.ErrNoRows:
		this.ErrorWithMsg (nil, w, http.StatusUnauthorized, cmd.ApiErrorCode_noIdentifiersForUser, "Email or password invalid")
		return
	case nil:
		// no return, just fall through
	default:
		this.ServerError (err, cmd.ApiErrorCode_dbError, w) // this is generically bad
		return
	}

	this.SuccessWithMsg (ctx, w, struct {
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
	
	this.UserError (ctx, nil, w, user) // we're done
}
