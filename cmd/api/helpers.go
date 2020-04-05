/*! \file helpers.go
	\brief Contains helper functions for incoming http requests
*/

package main

import (
	"github.com/NathanRThomas/boiler_api/cmd"
	"github.com/NathanRThomas/boiler_api/pkg/models"

	"github.com/pkg/errors"
			
	//"fmt"
	"net/http"
	"strings"
	"database/sql"
	"context"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief For the portal endpoints, most of them require an authenticated user.  This authenticates them and pass the info through the context
*/
func (this *app_c) bearerCheck (next http.Handler) http.Handler {
	return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		// parse out an user from what's passed in by the body
		ctx := r.Context()
		
		// see if we're logged in
		authToken := r.Header.Get ("Authorization")
		if len(authToken) == 0 { // no authorization
			this.ErrorWithMsg (nil, w, http.StatusUnauthorized, cmd.ApiErrorCode_noIdentifiersForUser, "Please login")
			return
		}

		bearerSplit := strings.Split (authToken, "Bearer")
		if len(bearerSplit) == 2 {
			authToken = strings.TrimSpace (bearerSplit[1]) // update this
		} // else let's assume it's missing the word "bearer" and it's just the user_id:token
		
		userSplit := strings.Split (authToken, ":") // split out our user_id:token
		if len(userSplit) != 2 { // invalid format
			this.ErrorWithMsg (nil, w, http.StatusUnauthorized, cmd.ApiErrorCode_noIdentifiersForUser, "Please login")
			return
		}

		user := &models.User_t{}	// setup our user object
		user.ID.Set (userSplit[0])
		user.Token.Set (userSplit[1])
		
		//see if this user is "good"
		err := this.Users.TokenLogin (user)

		switch errors.Cause (err) {
		case nil:
			// we have an user, so add them to the context and get our bot and org as well
			ctx = context.WithValue(ctx, "user", user)	// save our user in our context

			// now fire the next call with our user info now set
			next.ServeHTTP(w, r.WithContext (ctx))	// send it along

		case models.ErrType_noIdentifiers, sql.ErrNoRows: // we couldn't log in
			this.ErrorWithMsg (nil, w, http.StatusUnauthorized, cmd.ApiErrorCode_noIdentifiersForUser, "Please login")

		default: // something "bad" happened
			this.ServerError (err, cmd.ApiErrorCode_dbError, w)
		}
    })
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- QUERY PARAMETERS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//
