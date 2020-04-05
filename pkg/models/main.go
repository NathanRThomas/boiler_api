/*! \file main.go
	\brief Main stuff for redis
*/

package models 

import (
	"github.com/pkg/errors"

	"regexp"
	"strconv"
	"strings"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"database/sql"
	"net/http"
	"net/url"
	"encoding/json"
	"html/template"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type ProductionLevel int
const (
    ProductionLevel_Dev      	ProductionLevel = iota
    ProductionLevel_Production
    ProductionLevel_Staging
)

const salt = "something salty" //salt for passwords
var MaxMask int64 = 2 << 53 -1 // max int we can hold in javascript

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ERRORS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var (
	ErrType_userMissing 			= errors.New("User missing from context")
	ErrType_noIdentifiers 			= errors.New("Bearer token is missing identifiers")

	ErrType_invalidUUID 			= errors.New("Invalid UUID")
	ErrType_permission				= errors.New("You don't have permission to do this")
	
	ErrType_tookLongTime			= errors.New("request took a long time to complete")

	ErrType_nonFatal				= errors.New("non fatal error occured")
	ErrType_returnToUser 			= errors.New("Error Occured")
)


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TYPES -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type EnumString string

func (this EnumString) Int () int64 {
	id, _ := strconv.ParseInt(this.String(), 10, 64)
	return id
}

func (this EnumString) String () string {
	return string(this)
}

func (this *EnumString) Set (in string) {
	*this = EnumString(in)
}

// -----

type UUID string

func (this UUID) Valid () bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
    return r.MatchString(this.String())
}

func (this UUID) Int () int64 {
	id, _ := strconv.ParseInt(this.String(), 10, 64)
	return id
}

func (this UUID) String () string {
	return string(this)
}

func (this *UUID) Set (in string) {
	*this = UUID(in)
}

func (this *UUID) Nullable () (out sql.NullString) {
	if this.Valid() {
		out.String = this.String()
		out.Valid = true
	}
	return
}

func (this *UUID) Key (in string) string {
	return in + ":" + this.String()
}

/*! \brief This will try to set the id from the query param id
*/
func (this *UUID) Query (r *http.Request) bool {
	params := r.URL.Query()
	if len(params["id"]) > 0 && len(params["id"][0]) > 0 {
		this.Set (params["id"][0])
	}
	return this.Valid()
}

// -----
type ApiString string

func (this *ApiString) clean () { this.Set (strings.TrimSpace(string(*this))) }

func (this *ApiString) String () string { 
	this.clean()
	return string(*this) 
}

func (this *ApiString) Set (str string) { *this = ApiString(str) }

func (this *ApiString) Valid () bool {
	return len(this.String()) > 0
}

func (this *ApiString) Email () bool {
	local := this.String()
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

func (this *ApiString) Phone () bool {
	match := regexp.MustCompile(`^[\+1 -\(\)]*([2-9]\d{2})[\(\)\. -]{0,2}(\d{3})[\. -]?(\d{4})$`)
	resp := match.FindStringSubmatch(strings.TrimSpace(this.String()))
	if len(resp) == 4 {
		this.Set(fmt.Sprintf("1%s%s%s", resp[1], resp[2], resp[3])) //format this how we want it
		return true
	}
	return false // this is bad
}

func (this *ApiString) Password () bool {
	local := this.String()
	
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

func (this *ApiString) Hash () string {
	if !this.Valid() { return "" } // can't hash it if it's empty
	hash := sha256.Sum256([]byte(this.String()))
	return hex.EncodeToString(hash[:])
}

func (this *ApiString) PassRequires () string {
	return "Passwords must contain at least 8 characters and include a lower case and upper case letter"
}

/*! \brief Validates the string as a url and sets it if it is
*/
func (this *ApiString) Url () bool {
	resp, err := url.Parse (this.String())
	if err != nil { return false }

	if len(resp.Scheme) == 0 {	// malformed url
		resp, err = url.Parse ("http://" + this.String()) // add in the http
		if err != nil { return false }
	}

	if len(resp.Scheme) == 0 || len(resp.Host) == 0 { return false } // not a real url

	if strings.Index (resp.Host, ".") < 1 { return false } // we need a period in this

	this.Set (resp.String())
	return true
}

/*! \brief Sets the value of our object using an sprintf
*/
func (this *ApiString) Sprintf (msg string, params ...interface{}) {
	this.Set (fmt.Sprintf (msg, params...))
}

func (this *ApiString) Int () int64 {
	id, _ := strconv.ParseInt(this.String(), 10, 64)
	return id
}

func (this *ApiString) JsonString () string {
	jStr, err := json.Marshal(this.String())
	if err == nil {
		return string(jStr[1 : len(jStr)-1])
	}
	return ""
}

func (this *ApiString) Nullable () (out sql.NullString) {
	if this.Valid() {
		out.String = this.String()
		out.Valid = true
	}
	return
}

func (this *ApiString) Html () (template.HTML) {
	return template.HTML(this.String())
}

/*! \brief Preps a string for use in the database by removing "weird" characters
 */
 func (this *ApiString) SafeRegex () string {
	r := regexp.MustCompile("[^a-zA-Z0-9]")
	return strings.ToLower(r.ReplaceAllString(this.String(), ""))
}

/*! \brief Does a case insensitive string compare
*/
func (this *ApiString) Equal (in string) bool {
	return strings.EqualFold (this.String(), in)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

