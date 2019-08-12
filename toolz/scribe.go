/*! \file box.go
 *  \brief Contains non-class based functions that can be re-used everywhere
 */

package toolz

import (
    "time"
    "fmt"
    "log"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFINES -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const timebox_timeout   = 5000

const (
	DebugLevel_none 		= iota
	DebugLevel_all
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! Used for debugging/verifying timing with sections of code.  useful for debugging, but also for indicators that something is off
*/
type TimeBox_c struct {
    notes       []string
    stamps      []time.Time
}

func (this *TimeBox_c) Note (desc string, vals ...interface{}) {
    this.notes = append(this.notes, fmt.Sprintf(desc, vals...))
    this.stamps = append(this.stamps, time.Now())
}

func (this *TimeBox_c) Defer () {
    this.Note("exit") //mark our last time

    //now see if we have a long enough duration to care
    stampCnt := len(this.stamps)
    if stampCnt > 1 && this.stamps[stampCnt -1].Sub(this.stamps[0]) >= time.Millisecond * timebox_timeout {  //this took too long
        log.Printf("******************* TIMEBOX **************\n%s - %s\n", this.stamps[0], this.notes[0])
        for i := 1; i < stampCnt; i++ { //now do a timediff against the rest
            log.Printf("%s - %s\n", this.stamps[i].Sub(this.stamps[i-1]), this.notes[i])
        }
        log.Println("================== TIMEBOX ==============")
    }
}

/*! \brief Creates a timebox and seeds the first note
*/
func TimeBox (desc string, vals ...interface{}) *TimeBox_c {
    tb := &TimeBox_c {}
    tb.Note (desc, vals...)
    return tb
}

/*! Handles logging of different sections of the api entry point
*/
type Tutor struct {}

func (this Tutor) Log (debugLevel int64, params ...interface{}) {
	text := "DEBUG :: "
	for _, val := range params {
		text += fmt.Sprintf("%+v :: ", val)
	}

	switch debugLevel {
	case DebugLevel_all:
		log.Println(text)
	}
}