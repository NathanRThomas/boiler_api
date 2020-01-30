/*! \file main.go
	\brief Global level defines
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
	
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const (
    ProductionType_Dev      = iota
    ProductionType_Production
    ProductionType_Staging
)

const salt = "super salty" //salt for admin passwords

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ERRORS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var (
	ErrType_txMissing 				= errors.New("Database transaction bad")
	ErrType_userMissing 			= errors.New("User missing from context")

	ErrType_invalidUUID 			= errors.New("Invalid UUID")
	ErrType_noIdentifiers			= errors.New("Unable to identify")
	ErrType_platformNotFound		= errors.New("Platform not found")
	ErrType_permission				= errors.New("You don't have permission to do this")
	ErrType_cooldown				= errors.New("outside of an appropriate timeframe to send a message")
	ErrType_doNotInclude			= errors.New("Don't include this one")

	ErrType_botPlatformTaken		= errors.New("Platform for bot already in use")
	ErrType_languageNotFound		= errors.New("Language id not found")
	ErrType_tookLongTime			= errors.New("request took a long time to complete")

	ErrType_nonFatal				= errors.New("non fatal error occured")

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

// -----
type ApiString string

func (this *ApiString) clean () { this.Set (strings.TrimSpace(string(*this))) }

func (this *ApiString) Get () string { 
	this.clean()
	return string(*this) 
}

func (this *ApiString) Set (str string) { *this = ApiString(str) }

func (this *ApiString) Valid () bool {
	local := this.Get()
	if len(local) == 0 { return false }
	
	return true // we're good
}

// -----
type ApiEmail string

func (this *ApiEmail) clean () { this.Set (strings.TrimSpace(string(*this))) }

func (this *ApiEmail) Get () string { 
	this.clean()
	return string(*this) 
}

func (this *ApiEmail) Set (str string) { *this = ApiEmail(str) }

func (this *ApiEmail) Valid () bool {
	local := this.Get()
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

// -----
type ApiPassword string
func (this *ApiPassword) Get () string { return string(*this) }
func (this *ApiPassword) Set (str string) { *this = ApiPassword(str) }

func (this *ApiPassword) Valid () bool {
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

func (this *ApiPassword) Hash () string {
	return hashIt (this.Get())
}

// -----
type ApiPhone string
func (this *ApiPhone) Get () string { return string(*this) }
func (this *ApiPhone) Set (str string) { *this = ApiPhone(str) }

func (this *ApiPhone) Valid () bool {
	match := regexp.MustCompile(`^[\+1 -\(\)]*([2-9]\d{2})[\(\)\. -]{0,2}(\d{3})[\. -]?(\d{4})$`)
	resp := match.FindStringSubmatch(strings.TrimSpace(this.Get()))
	if len(resp) == 4 {
		this.Set(fmt.Sprintf("1%s%s%s", resp[1], resp[2], resp[3])) //format this how we want it
		return true
	}
	return false // this is bad
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


func hashIt (in string) string {
	hash := sha256.Sum256([]byte(in))
    return hex.EncodeToString(hash[:])
}
