/*! \file user.go
	\brief user related objects
*/

package models 

import (
	//"fmt"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type UserMask int64
const (
	UserMask_deleted			UserMask = 1 << iota  //
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DATA STRUCTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type User_t struct {
	ID UUID
	Email, Password, Token ApiString
	Mask UserMask `json:",omitempty"`
	Created time.Time
	Attr struct {
		First, Last string `json:",omitempty"`
	}
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *User_t) SetToken () {
	this.Token.Set (this.Password.String() + this.Email.String() + salt)
	this.Token.Set(this.Token.Hash())
}