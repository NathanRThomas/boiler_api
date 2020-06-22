/*! \file helpers.go
	\brief Re-usable code and functions used by the web main code
*/

package cmd

import (
	"github.com/NathanRThomas/boiler_api/pkg/models"

	"github.com/pkg/errors"

	"fmt"
	"net/http"
	"context"
	"encoding/json"
	"database/sql"
	"math/rand"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONST -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


//! Basic response object that we send back when there's nothing else to be said
type ApiError_t struct {
	Error struct {
		Msg       string
		Code      int
	}
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HANDLERS ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App_c) RandError () string {
	terms := [...]string {"There was an issue with the monkies", "Houston we have a problem", "The gerbil stopped running",
		"Looks like the Fatherboard is burned out", "There's a loose wire between the mouse and keyboard", "That really didn't work",
		"If at first you don't succeed... give up, this is a deterministic system", "The monkies are out of their cages",
		"I can't tell for sure but I'm pretty sure this is your fault", "I think there's a gas leak", "Russia is hacking us", 
		"North Korea is attacking", "We're being hacked by China" }

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := rnd.Intn(len(terms) - 1)
	return terms[n]
}

/*! \brief Handles pulling in data from our body into whatever object we need to read it into
*/
func (this *App_c) ParseFromBody (ctx context.Context, out interface{}) error {
	if body, ok := ctx.Value("body").([]byte); ok && len(body) > 0 { 
		return errors.WithStack (json.Unmarshal (body, out))
	}
	return nil
}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (this *App_c) ServerError (err error, code int, w http.ResponseWriter) {
	this.ErrorWithMsg (err, w, http.StatusInternalServerError, code, this.RandError())
}

/*! \brief Handles a "successful" api call
	Commits the transaction log, and writs out our response to the user
*/
func (this *App_c) SuccessWithMsg (w http.ResponseWriter, in interface{}) {
	if in == nil { w.Write([]byte("{}")); return } // we're done

	jOut, err := json.Marshal (in)
	if err == nil {
		w.Write(jOut)	// finally everything is good and we can respond to the user
		return 
	}
	err = errors.Wrapf (err, "%+v\n", in) // i want to know if this is failing
	this.ServerError (err, ApiErrorCode_jsonMarshal, w)
}

/*! \brief Main error handling function, doesn't do anything with the transaction, but handles the error response object
*/
func (this *App_c) ErrorWithMsg (err error, w http.ResponseWriter, httpStatus, code int, msg string, params ...interface{}) {
	if err != nil { this.StackTrace (err) }  // record this

	final := fmt.Sprintf(msg, params...)
	if len(final) == 0 { final = http.StatusText(httpStatus) }	// default to the text version of the status code

	errT := ApiError_t {}
	errT.Error.Msg = final 
	errT.Error.Code = code
	jOut, err := json.Marshal (errT)	// always use this object for errors
	if err != nil { this.StackTrace (err) }  // record this

	http.Error (w, string(jOut), httpStatus)	// now give the requester some info
}

/*! \brief I seemed to be calling this a lot, so i put a wrapper around missing/bad url and query params
*/
func (this *App_c) MissingParam (w http.ResponseWriter, msg string, params ...interface{}) {
	this.ErrorWithMsg (nil, w, http.StatusBadRequest, ApiErrorCode_invalidInputField, msg, params...)
}

func (this *App_c) Forbidden (w http.ResponseWriter, msg string, params ...interface{}) {
	this.ErrorWithMsg (nil, w, http.StatusForbidden, ApiErrorCode_permissions, msg, params...)
}

/*! \brief When we had an issue that may have been caused by user input, this decides which error should be returned to the user
*/
func (this *App_c) Respond (err error, w http.ResponseWriter, success interface{}) {
	switch errors.Cause(err) {
	case models.ErrType_returnToUser:
		this.ErrorWithMsg (nil, w, http.StatusBadRequest, ApiErrorCode_invalidInputField, err.Error())
	
	case models.ErrType_permission:
		this.ErrorWithMsg (nil, w, http.StatusForbidden, ApiErrorCode_internal, "You don't have access to this")
	
	case sql.ErrNoRows: // expected 404 error
		w.WriteHeader(http.StatusNotFound)
	
	case nil:
		this.SuccessWithMsg (w, success) // everything was good
	
	default:
		this.ServerError (err, ApiErrorCode_internal, w)
	}
}

