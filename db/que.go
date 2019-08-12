/*! \file que.go
 *  \brief Pulled out the specific slave/master queueing functions
 */

package db

import (
    //"fmt"
	"encoding/json"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const queenQueKey		= "task:queen"
const workerQueKey		= "task:worker"

const (
	WorkerQue_nothing			= iota
	
)

const (
	QueenQue_nothing			= iota

)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC STRUCTS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type WorkerQue struct {
	Type, MiscInt int64 `json:",omitempty"`
    MiscString string `json:",omitempty"`
}

type QueenQue struct {
	Type, MiscInt int64 `json:",omitempty"`
    MiscString string `json:",omitempty"`
}

type Que_c struct { }

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MASTER FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief For when a slave task needs to pass something back to the master
 */
func (this *Que_c) MasterTask (que QueenQue) {
    j, _ := json.Marshal(que)
    Redis.Lpush (redis_index_tasks, queenQueKey, string(j[:]))
}

/*! \brief Gets the next item from the list for the master que to do
 */
func (this *Que_c) PopMasterQue () (*QueenQue) {
	if rs := Redis.Rpop(redis_index_tasks, queenQueKey); rs != nil { 
        que := &QueenQue{}
        UM(rs, que) 
        return que
    }
	return nil
}

/*! \brief Just looks to see if we may have a problem with the master que based on its length
*/
func (this *Que_c) MasterQueSize () bool {
    if Redis.Llen(redis_index_tasks, queenQueKey) > 100 { return true }
    return false
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- SLAVE FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Handles scheduling a slave que task
 */
func (this *Que_c) WorkerTask (que WorkerQue) {
    j, _ := json.Marshal(que)
    Redis.Lpush (redis_index_tasks, workerQueKey, string(j[:]))
}

/*! \brief Pops the next slave task off the que for processing
 */
func (this *Que_c) PopWorkerQue () (*WorkerQue) {
	if rs := Redis.Rpop(redis_index_tasks, workerQueKey); rs != nil { 
        que := &WorkerQue{}
        UM(rs, que)
        return que
    }
	return nil
}

func (this *Que_c) WorkerQueFull () bool {
    if Redis.Llen(redis_index_tasks, workerQueKey) > 20000 { return true }
    return false
}
