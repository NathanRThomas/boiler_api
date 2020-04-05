/*! \file queen.go
	\brief Contains all functions related to the queen tasks
*/

package main

import (
	"github.com/NathanRThomas/boiler_api/cmd"
	"github.com/NathanRThomas/boiler_api/pkg/models"
	
	"github.com/pkg/errors"
		
	//"fmt"
	"context"
	"database/sql"
	"time"
	"runtime"
	"runtime/debug"

)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type taskFunc func (context.Context, chan error)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HELPER FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOW LEVEL ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOCAL FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- SCHEDULES ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Checks out the schedules table for tasks that need to be done
*/
func (this *app_c) doSchedules (ctx context.Context, ch chan error) {
	next, err := this.tasks.NextSchedule () // get the next scheduled task
	if next == nil || err != nil {
		if errors.Cause (err) == sql.ErrNoRows { err = nil }
		ch <- err
		return // we're done
	}

	switch next.Type {
	default:
		ch <- errors.Wrapf (models.ErrType_nonFatal, "unknown schedule type :%d", next.Type)
	case models.ScheduleType_none:
		ch <- errors.Wrap (models.ErrType_nonFatal, "got a schedule with no type")
	}
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MESSAGES ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Wrapper around each individual function above
	We want to create a new timeout for each task
*/
func (this *app_c) startQueenFunc (ctx context.Context, fn taskFunc) (err error) {
	ctx, cancel := context.WithTimeout (ctx, time.Second * cmd.ContextTimeout) // no single task should take longer than this, otherwise we have an issue
	defer cancel()

	ch := make(chan error, 1)	// channel for tracking when the entry call finishes

	go func() {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				this.ErrorLog.Println(err)
				debug.PrintStack()
				// ch <- nil i'm actually going to let this timeout, if there is a panic happening we don't want to just fill the logs with it
			}
		}()

		fn (ctx, ch)	// handle things
	}()
	
	select {
	case <- ctx.Done():
		//this is bad, the context expired on us
		err = errors.Errorf ("context expired for queen: %s\n", ctx.Err())
	case err = <- ch: // finished normally
	}
	
	return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Queues background queen tasks
	This should only be run by one thread across the whole cluster.  These are for single non-thread safe tasks
*/
func (this *app_c) queen () {
	defer this.WG.Done() //so our main thread can move on when this finishes

	ctx := context.WithValue (context.Background(), "slackConfig", &cmd.CFG.Slack)	// add this to our context, some tasks need it
	ctx = context.WithValue (ctx, "mailgunConfig", &cmd.CFG.Mailgun)	// add this to our context, some tasks need it
	
	cnt := 0
    for this.Running {
		
		this.StackTrace (this.startQueenFunc (ctx, this.doSchedules))	// handle our scheduled re-curring tasks
		
		if cnt >= 10 { // these don't have to run as frequently "low-level" tasks
			
			cnt = 0
		} else {
			cnt++
		}

		for x := 0; x < 5; x++ {
			if !this.Running { return }     //we've stopped our thread early
			runtime.Gosched()
            time.Sleep(time.Second)
        }
	}
}
