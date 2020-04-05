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

//----- TASKS -----//
type QueTask int
const (
	QueTask_nothing 			QueTask = iota 
	
)

//----- SCHEDULES -----//
// these are from the schedules table in the database
type ScheduleType int64
const (
	ScheduleType_none 				ScheduleType = iota
	
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DATA STRUCTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Que_t struct {
	Type QueTask
	Expires int64
    UserID UUID `json:",omitempty"`
}

type Schedule_t struct {
	ID UUID `json:",omitempty"`
	Type ScheduleType
	Interval ApiString
	Attr struct {
		Desc, Email string `json:",omitempty"`
	}
}
