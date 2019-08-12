/*! \file import.go
 *  \brief All things related to saving of files
 *
 */

package web

import (
    //"fmt"
	"encoding/json"
	"net/http"
	
	"github.com/NathanRThomas/boiler_api/toolz"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOCAL CLASS -------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! Main class for the user file
 */
type Import_c struct {
	R *http.Request
    fc      toolz.FileCabnet_c
}

//--- PRIVATE MEMBERS ------------------------------------------------------------------------------------------------//

/*! \brief Handles saving of an uploaded file
*/
func (this *Import_c) fileUpload (formTarget, formID string, compress bool) ([]byte, error) {
	url := ""
	file, _, err := this.R.FormFile(formTarget)
	
 	if err == nil {
		url, err = this.fc.LocalUpload(file, formID, compress, false)
		file.Close()
	}
	
	if err == nil {
		return json.Marshal(struct {
			APIResponse_t
			Url		string	`json:"url"`
		}{APIResponse_t{true, formTarget + " uploaded successfully!", api_no_error}, url})
	} else {
		return json.Marshal(APIResponse_t{false, err.Error(), api_no_error})
	}
}



  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point for the user section of the api.  Any url that starts with "user" goes here
 */
func (this *Import_c) Entry(statement string, r *http.Request) ([]byte, error) {
	
	switch statement {
	case "image":
		return this.fileUpload(statement, "", true)
	
	}
	
	return json.Marshal(APIResponse_t{false, "Unknown call", api_error_unknown_call})
}
