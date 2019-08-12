/*! \file compass.go
 *  \brief Used for redirection and meta data handling.
 */

package web

import (
	//"fmt"
    "strings"
    
	"github.com/NathanRThomas/boiler_api/db"
	"github.com/NathanRThomas/boiler_api/toolz"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Compass_c struct {
	ip 		    toolz.IP_c
	users       db.User_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Entry for when we're dealing with a user level url redirect
*/
func (this *Compass_c) link(part, id, primary, secondary string) (finalUrl, finalHtml string) {
	switch (strings.ToLower(part)) {
	
	case "follow":
		//user := &db.User_t { ID : id }
		
	case "url": //pre-generated url that stays static
		
	}

    return //we're done
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point for the tag section of the api.  Any url that starts with "tag" goes here
 */
func (this *Compass_c) Entry(parts []string) (string, string) {
    switch (strings.ToLower(parts[0])) {
	
	case "l": //pre-generated url that stays static
		
	}

    return "", "" //we're done
}
