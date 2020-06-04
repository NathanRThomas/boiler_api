/*! \file mailgun.go
 *  \brief Class for handling sending of emails
 */

package toolz

import (

	"github.com/mailgun/mailgun-go/v3"
	"github.com/pkg/errors"

    //"fmt"
    "regexp"
	"strings"
	"context"
    "time"
    
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const mailgun_default_from  = "Admin<info@example.com>"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type MailgunConfig_t struct {
	Domain, Key string
}

type Mailgun_c struct {
    
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Sends a single plain text email to a user
 */
func (this *Mailgun_c) Send (ctx context.Context, from, subject, body, html, campaign string, to ...string) error {
	config, ok := ctx.Value ("mailgunConfig").(*MailgunConfig_t) // get our config
	if !ok { return errors.New ("mailgun config missing from context") }

	gun := mailgun.NewMailgun (config.Domain, config.Key)

    if len(from) < 1 { from = mailgun_default_from }
    
	email := gun.NewMessage(from, subject, body, to...)   //Start the email
	
	if len(html) > 0 { 
		email.SetHtml(html) 
	} else {
		email.SetTracking (false) // if it's not html, then we don't want to track it
	}
    
	if len(campaign) > 0 { email.AddTag(campaign) }
	
	gunCtx, cancel := context.WithTimeout (ctx, time.Second * 20) // give 20 seconds for this task
	defer cancel()
    
	_, _, err := gun.Send(gunCtx, email)   //send the message
	return errors.Wrap (err, subject)
}

/*! \brief Does a validate call against the mailgun server
 */
func (this *Mailgun_c) ValidateEmail (email *string) bool {
    *email = strings.Trim(*email, " ")
    if len(*email) > 3 {
        if match, _ := regexp.MatchString("^.+@.+\\..+$", *email); match {
			if strings.Index(*email, " ") < 0 {  //no spaces
				return true
			}
			/*  this is probably overkill right now
			v1, _ := m.gun.ValidateEmail (email)
			return v1.IsValid
			*/
        }
    }
    return false
}