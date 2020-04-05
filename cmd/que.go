/*! \file que.go
	\brief Shared queing functions, specifically for message ques
*/

package cmd 

 import (
	"github.com/NathanRThomas/boiler_api/pkg/models"
	
	"github.com/pkg/errors"
	
	//"fmt"
	"time"
	"context"

 )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFINES -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const taskThreadCount		= 9 // Number of threads in our "pool" of task handlers

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOCAL FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

//----- MAIN ENTRY -----//

/*! \brief Publically avialable entry point into this shared class
	Looks for a que object and handles whatever is thrown at it
*/
func (this *App_c) TaskQueEntry (ctx context.Context, ch chan error, que *models.Que_t) {
	// do some base-work here
	var user *models.User_t
	
	if que.UserID.Valid() {
		user = &models.User_t { ID: que.UserID }	// init this
		err := this.Users.Get (user) // get our user
		if err != nil { ch <- err; return }

		ctx = context.WithValue (ctx, "user", user)	// add this to our context
	}

	// now see what our switch is doing
	switch que.Type {
	/*
	case models.QueType_notifyUser:
		if user == nil { ch <- errors.Errorf("user is missing"); return }
	*/

	default:
		ch <- errors.Errorf("Unknown Que Type : %d", que.Type)
		return
	}

	// if we're here we're done
	ch <- nil
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief We have lots of "things" that we need to que for completion in a background process.  
			These items stay locally in memory for this instance, so it's important it never gets too large and that it completes before
			the service terminates
*/
func (this *App_c) StartTaskQue () {
	this.TaskQue = make (chan *models.Que_t, models.MaxQueSize)	// allocate our global channel

	// launch our background proccesing thread
	for i := 0; i < taskThreadCount; i++ {	// this creates n processing threads using the anonymous function below
		this.WG.Add(1)
		go func () {
			defer this.WG.Done()
			ch := make(chan error, 1)	//channel for tracking when the entry call finishes

			for {	// stay in this loop. as long as we're still running or there's still messages in the que, we're not done
				select {
				case que := <- this.TaskQue:
					if que == nil { return } // means the channel was closed
					ctx, cancel := context.WithTimeout (context.Background(), time.Second * ContextTimeout) // no single task should take longer than this, otherwise we have an issue

					ctx = context.WithValue (ctx, "slackConfig", &CFG.Slack)	// add this to our context, some tasks need it
					ctx = context.WithValue (ctx, "mailgunConfig", &CFG.Mailgun)	// add this to our context, some tasks need it

					//we got a message in our que
					go this.TaskQueEntry (ctx, ch, que)	// handle things

					var err error
					select {
					case <-ctx.Done():
						//this is bad, the context expired on us
						err = errors.Errorf ("context expired for que: %s : %+v\n", ctx.Err(), que)
					case err = <- ch: // finished normally
					}

					this.StackTrace (err) // record this error, if one exists

					cancel()	// don't defer since we're in a loop, just call it here everytime
				}
			}
		}()
	}
}
