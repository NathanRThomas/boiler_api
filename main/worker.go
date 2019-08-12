/*! \file slave.go
 *  \brief Class for handling slave processing of things.  you can run multiple instances of this
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

const maxSlaveWorkers   = 2

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type worker_c struct {
	wg      		*sync.WaitGroup
	running 		*bool
    que             db.Que_c
    aws             toolz.AWS_c
    fc              toolz.FileCabnet_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief This is our worker thread.  This actually handles things in parallel for us
 */
func (this *worker_c) worker (que <-chan *db.WorkerQue, wg *sync.WaitGroup) {
    defer wg.Done()
    for q := range que {    //just keep looping through this channel while it remains open
        switch q.Type {
            
            default:
                toolz.Err("Slave Que type not found: %d", q.Type)
        }
    }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point.  This looks through the agent and pulls the next task to do
 */
func (this *worker_c) Tasks () {
    this.wg.Add(1)
    defer this.wg.Done() //so our main thread can move on when this finishes
    locWait := new(sync.WaitGroup)
    locWait.Add(maxSlaveWorkers) 
    
    que := make(chan *db.WorkerQue, maxSlaveWorkers) //this channel is where we're going to stick our tasks we've pulled off the redis que
    for w := 0; w < maxSlaveWorkers; w++ {
        go this.worker(que, locWait)
    }
    
    for *this.running {
        didSomething := false
        
        //peal the next user of the subscription list
        q := this.que.PopWorkerQue()
        if q != nil && q.Type != db.WorkerQue_nothing {
            que <- q    //add this to our task list
            didSomething = true
        }
        
        if !didSomething { time.Sleep(time.Millisecond * 250) } // hang out, nothing to do
    }
    
    close(que)  //we're done, tell the other threads to stop too
    for len(que) > 0 { time.Sleep(time.Millisecond * 100) }	// wait here until the que is empty

    locWait.Wait()  //wait for our slaves to finish
}