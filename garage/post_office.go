/*! \file post_office.go
 *  \brief Class for handling the high level messaging, emails etc
 */

package garage

import (
    "fmt"
    "bytes"
    "html/template"
    "os"
    
    "github.com/NathanRThomas/boiler_api/db"
    "github.com/NathanRThomas/boiler_api/toolz"
    
    "github.com/NathanRThomas/plivo-go"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type PostOffice_c struct {
    mailman     toolz.Mailman_c
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Generates the template of interest
*/
func this *PostOffice_c) genTemplate (name string, data interface{}) string {
    buf := new(bytes.Buffer)
    toolz.ErrChk(template.Must(template.ParseFiles(os.Getenv("API_TEMPLATE") + "email/" + name)).Execute(buf, data))
    return buf.String()
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func this *PostOffice_c) ForgotPassword (email string) {
	go this.mailman.Send("", "Password Reset",
		fmt.Sprintf("Sorry you're having problems logging in."), "", "", email)
}

/*! \brief Welcome email when a user first creates an account on the bot
*/
func this *PostOffice_c) Welcome (email string) {
    if this.mailman.ValidateEmail(&email) {    //we have a valid email to send this to
        go this.mailman.Send("", "Welcome!",
            fmt.Sprintf("Hey"), 
            this.genTemplate("welcome.html", nil), "Welcome", email)
    }
}
