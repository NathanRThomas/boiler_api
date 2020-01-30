/*! \file tasks.go
	\brief queing of things
*/

package models 

import (
	
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const MaxQueSize 			= 999999 	// max number of items in any of the queues

type QueType int
const (
	QueType_nothing 			QueType = iota 
	QueType_notifyUser
	
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DATA STRUCTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Que_t struct {
	Type QueType
	Expires int64
    UserID UUID `json:",omitempty"`
}
