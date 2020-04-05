/*! \file task.go
	\brief Cockroach specific to the tasks table
	These are one-time things that need to happen at a specific time

*/

package cockroach

import (
	"github.com/NathanRThomas/boiler_api/pkg/models"

	"github.com/pkg/errors"

	"fmt"
	"time"
)

type Task_c struct {
	toolz_c
}
  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- SCHEDULES ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Grabs the single next schedule that should be executed
	Will return nil for the schedule and the error if there's nothing to do
*/
func (this *Task_c) NextSchedule () (*models.Schedule_t, error) {
	next := &models.Schedule_t{}
	var jAttr []byte

	err := db.QueryRow (`SELECT id, schedule_type, attrs, interval FROM schedules WHERE next_date < $1 ORDER BY next_date LIMIT 1`, time.Now()).
		Scan(&next.ID, &next.Type, &jAttr, &next.Interval)
	
	if err != nil { return nil, errors.WithStack (err) }

	err = this.UM (jAttr, &next.Attr)
	if err != nil { return nil, err } // bail

	// now try to schedule this one in the future
	interval := "10 year" // default to way in the future, so we don't keep picking a "bad" one up
	if !next.Interval.Valid() { 
		// update this so we don't keep trying to re-send it
		// here's the deal, if the above update fails i want to know, otherwise return a non-fatal error so that the above update still gets committed
		err = errors.Wrapf (models.ErrType_nonFatal, "%s : schedule could not be set to a new interval", next.ID) 
	} else {
		interval = next.Interval.String() // just pull this out
	}

	lErr := this.Exec (fmt.Sprintf(`UPDATE schedules SET next_date = next_date + INTERVAL '%s' WHERE id = $1`, interval), next.ID)
	if lErr != nil { return nil, lErr } // the update failed, so bail hard here

	return next, err // return our "non-fatal" error from above, do this so the update to the future always works
}
