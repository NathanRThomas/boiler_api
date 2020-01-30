/*! \file que.go
	\brief Shared queing functions, specifically for message ques

	Created 2019-11-12 by NateDogg
*/

package cmd 

 import (
	"github.com/NathanRThomas/boiler_api/pkg/models"
	"github.com/NathanRThomas/boiler_api/pkg/toolz"
	
	"github.com/pkg/errors"
	
	//"fmt"
	"time"
	"context"

 )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFINES -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const taskThreadCount		= 3 // Number of threads in our "pool" of task handlers

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOCAL FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

//----- MAIN ENTRY -----//

/*! \brief Publically avialable entry point into this shared class
	Looks for a que object and handles whatever is thrown at it
*/
func (this *shared_c) Entry (ctx context.Context, ch chan bool, que *models.Que_t) {
	defer func () { ch<-true }()	// for when we're done

	ctx, tx, err := this.DB.Begin (ctx)	// create a transaction for this process
	if err != nil { this.StackTrace (err); return }
	
	// do some base-work here
	var user *models.User_t
	
	if que.UserID.Valid() {
		user = &models.User_t { ID: que.UserID }	// init this
		err = this.users.Get (tx, user) // get our user
		if err != nil { this.StackTrace (err); return }

		ctx = context.WithValue (ctx, "user", user)	// add this to our context
	}

	// now see what our switch is doing
	switch que.Type {
	case models.QueType_notifyUser:
		if user == nil { this.StackTrace (errors.Errorf("user is missing")); return }
		
	default:
		err = errors.Errorf("Unknown Que Type : %d", que.Type)
	}

	this.StackTrace (this.DB.Finish (ctx, tx, err))
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief This is used by the api and cli to handle sending of messages
	 the goal is to leave this channel open until all the tasks that could be putting messages into the que are done,
	 then we wait for this que to empty and we bail
*/
func NewQue (handle Handler_i, slackCFG *toolz.SlackConfig_t, mailgunCFG *toolz.MailgunConfig_t) (chan<- *models.Que_t) {
	_, errorLog := handle.GetLogs ()
	queChan := make(chan *models.Que_t, models.MaxQueSize)	// allocate our global channel

	//local instance of the que class for handling things
	class := shared_c { App_c: App_c {
			ProductionLevel: handle.GetProductionLevel(),
			ErrorLog: errorLog,
		},
		self: queChan,
	}

	class.Redis, class.DB, class.Cache = handle.GetDatabases ()

	wg := handle.GetWaitGroup ()
	// launch our background proccesing thread
	for i := 0; i < taskThreadCount; i++ {	// this creates n processing threads using the anonymous function below
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan bool, 1)	//channel for tracking when the entry call finishes

			for {	// stay in this loop. as long as we're still running or there's still messages in the que, we're not done
				select {
				case que := <-queChan:
					if que == nil { return } // means the channel was closed
					ctx, cancel := context.WithTimeout (context.Background(), time.Second * ContextTimeout) // no single task should take longer than this, otherwise we have an issue

					ctx = context.WithValue (ctx, "slackConfig", slackCFG)	// add this to our context, some tasks need it
					ctx = context.WithValue (ctx, "mailgunConfig", mailgunCFG)	// add this to our context, some tasks need it

					//we got a message in our que
					go class.Entry (ctx, ch, que)	// handle things

					select {
					case <-ctx.Done():
						//this is bad, the context expired on us
						errorLog.Printf ("context expired for que: %s : %+v\n", ctx.Err(), que)
					case <-ch: // finished normally
					}

					cancel()	// don't defer since we're in a loop, just call it here everytime
				}
			}
		}()
	}

	return queChan
}
