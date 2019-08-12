/*! \file master.go
 *  \brief Class for handling master processing of things.  There should only be one of these running at any time
 */

package main

import (
    //"fmt"
    "time"
    "sync"

    "github.com/NathanRThomas/boiler_api/db"
    "github.com/NathanRThomas/boiler_api/toolz"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const hour_wait_offset      = 5  //seconds to wait between each 1 hour task

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type queen_c struct {
	wg          	*sync.WaitGroup
	running 		*bool
    startTime       time.Time
	que             db.Que_c
	minute 			db.Worker_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Rips through the qued list of tasks for the master thread
 */
func (this *queen_c) masterQue () bool {
    for !this.timeCheck() {
        q := this.que.PopMasterQue()
        if q == nil { return true }  //we're done
        switch q.Type {
            
            default:
                toolz.Err("Unknown master que type: %d", q.Type)
        }
    }
    return false    //if this is our exit point, we didnt' get to finish everything in the que
}

/*! \brief Just makes sure we haven't exceeded our 1 minute time period for this
*/
func (this *queen_c) timeCheck () bool {
    if time.Since(this.startTime).Seconds() > 55 {
        db.Err("We ran out of time before calling the master que")  //record the error
        return true
    } else {
        return false
    }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief This function gets called once a minute, including right when the service starts
 */
func (this *queen_c) OneMinute () {
    this.wg.Add(1)
    defer this.wg.Done() //so our main thread can move on when this finishes

    //we need to make sure this finishes within a minute, as it's bad if we call the OneMinute function again before this finishes
    this.startTime = time.Now()

    if this.timeCheck() { return } //over time
	
	this.masterQue() //finish things off by going through the master que
}

/*! \brief This function gets called once an hour, and once right when the service starts
**/
func (this *queen_c) OneHour () {
    this.wg.Add(1)
    defer this.wg.Done() //so our main thread can move on when this finishes

    for x := 0; x < 14; x++ {    //loop here a little to give the one minute thread a chance to start
        time.Sleep(time.Second * hour_wait_offset)
        if !*this.running { return }     //we've stopped our thread early
    }
    
    if !*this.running { return }     //we've stopped our thread early

    time.Sleep(time.Second * hour_wait_offset)
	
	
}
