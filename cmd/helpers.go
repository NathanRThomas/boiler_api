/*! \file helpers.go
	\brief Re-usable code and functions used by the web main code

	Created 2019-10-24 by NateDogg
*/

package cmd

import (
	"github.com/NathanRThomas/boiler_api/pkg/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/auth0/go-jwt-middleware"
	"github.com/pkg/errors"

	"fmt"
	"net/http"
	"context"
	"strings"
	"regexp"
	"encoding/json"
	"database/sql"
	"math/rand"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- JWT ---------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App_c) CreateJWT (id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
		"exp": time.Now().Add(time.Minute * 1).Unix(),
	})
	
	// Sign and get the complete encoded token as a string using the secret
	str, err := token.SignedString([]byte(JWT_secret_key))
	return str, errors.WithStack (err)
}

func (this *App_c) GetJWTMiddleware () *jwtmiddleware.JWTMiddleware  {
	return jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(JWT_secret_key), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
		UserProperty: "jwt",
	})
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HANDLERS ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App_c) RandError () string {
	terms := [...]string {"There was an issue with the monkies", "Houston we have a problem", "The gerbil stopped running",
		"Looks like the Fatherboard is burned out", "There's a loose wire between the mouse and keyboard", "That really didn't work",
		"If at first you don't succeed... give up, this is a deterministic system", "The monkies are out of their cages",
		"I can't tell for sure but I'm pretty sure this is your fault" }

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
	return errors.WithStack (ErrType_bodyNotFound) // we were expecting this
}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (this *App_c) ServerError (err error, w http.ResponseWriter) {
	this.ErrorWithMsg (err, w, http.StatusInternalServerError, this.RandError())
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (this *App_c) NotFound (w http.ResponseWriter) {
	this.ErrorWithMsg (nil, w, http.StatusNotFound, "")
}
/*! \brief Sometimes we want to return an "error" but still want a 200 response code
*/
func (this *App_c) NormalError (ctx context.Context, w http.ResponseWriter, msg string) {
	this.SuccessWithMsg (ctx, w, msg)
}

/*! \brief Just a simple wrapper around the successWithMsg fuction for when we just have a message want to return
*/
func (this *App_c) ClientSuccess (ctx context.Context, w http.ResponseWriter, msg string) {
	this.SuccessWithMsg (ctx, w, msg)
}

/*! \brief Handles a "successful" api call
	Commits the transaction log, and writs out our response to the user
*/
func (this *App_c) SuccessWithMsg (ctx context.Context, w http.ResponseWriter, msg string) {
	var err error
	tx, ok := ctx.Value("tx").(*sql.Tx) // get our current db transaction
	if ok { 
		err = this.DB.Finish (ctx, tx, nil)	//commit the transaction

		if err == nil {
			w.Write([]byte(msg))	// finally everything is good and we can respond to the user
			return 
		}
	} else {
		err = errors.WithStack (models.ErrType_txMissing)
	}

	this.ServerError (err, w)
}

/*! \brief Main error handling function, doesn't do anything with the transaction, but handles the error response object
*/
func (this *App_c) ErrorWithMsg (err error, w http.ResponseWriter, status int, msg string, params ...interface{}) {
	if err != nil { this.StackTrace (err) }  // record this

	final := fmt.Sprintf(msg, params...)
	if len(final) == 0 { final = http.StatusText(status) }	// default to the text version of the status code

	http.Error (w, final, status)	// now give the requester some info
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TYPES -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var (
	ErrType_bodyNotFound		= errors.New("Body not found in context")
)

type InputString string
type InputEmail string
type InputPassword string
type InputPhone string

func (this *InputString) Get () string { return string(*this) }
func (this *InputString) Set (str string) { *this = InputString(str) }

func (this *InputString) Valid () bool {
	local := strings.TrimSpace(this.Get())
	if len(local) == 0 { return false }
	
	this.Set (local) // copy this back out
	return true // we're good
}

func (this *InputEmail) Get () string { return string(*this) }
func (this *InputEmail) Set (str string) { *this = InputEmail(str) }

func (this *InputEmail) Valid () bool {
	local := strings.TrimSpace(this.Get())
	if len(local) > 3 {
        if match, _ := regexp.MatchString("^.+@.+\\..+$", local); match {
			if strings.Index(local, " ") < 0 {  //no spaces
				this.Set (local) // copy this back out
				return true	// valid email
			}
        }
	}
	
	return false // not good
}

func (this *InputPassword) Get () string { return string(*this) }

func (this *InputPassword) Valid () bool {
	local := this.Get()
	
	matches := []string{"[a-z]", "[A-Z]"}
	if len(local) < 8 { return false }

	for _, m := range matches {
		r, _ := regexp.Compile(m)
		if len(r.FindAllString(local, 1)) == 0 { //ensure we have at least one match
			return false
		}
	}

	return true
}

func (this *InputPhone) Get () string { return string(*this) }
func (this *InputPhone) Set (str string) { *this = InputPhone(str) }

func (this *InputPhone) Valid () bool {
	match := regexp.MustCompile(`^[\+1 -\(\)]*([2-9]\d{2})[\(\)\. -]{0,2}(\d{3})[\. -]?(\d{4})$`)
	resp := match.FindStringSubmatch(strings.TrimSpace(this.Get()))
	if len(resp) == 4 {
		this.Set(fmt.Sprintf("1%s%s%s", resp[1], resp[2], resp[3])) //format this how we want it
		return true
	}
	return false // this is bad
}

