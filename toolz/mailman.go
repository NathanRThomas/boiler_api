/*! \file mailman.go
 *  \brief Class for handling sending of messages to users triggered by this api
 */

package toolz

import (
	//"fmt"
	"context"
    "regexp"
    "strings"
    
    "github.com/mailgun/mailgun-go"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type mailGunConfig_t struct {
	Domain, Key, Public, From string
}

type Mailman_c struct {
    inited  bool
    gun     mailgun.Mailgun
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Inits the Mailgun object based on our config settings
*/
func (this *Mailman_c) init () {
    if !this.inited {
        this.gun = mailgun.NewMailgun(AppConfig.MailGun.Domain, AppConfig.MailGun.Key)
        this.inited = true
    }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Sends a single plain text email to a user
 */
func (this *Mailman_c) Send (from, subject, body, html, campaign string, to ...string) error {
    this.init()
    if len(from) < 1 { from = AppConfig.MailGun.From }
    
    email := this.gun.NewMessage(from, subject, body, to...)   //Start the email
    
    if len(html) > 0 { email.SetHtml(html) }
    if len(campaign) > 0 { email.AddTag(campaign) }
    
    _, _, err := this.gun.Send(context.Background(), email)   //send the message
    return ErrChk(err)
}

/*! \brief Does a validate call against the mailgun server
 */
func (this *Mailman_c) ValidateEmail (email *string) bool {
    *email = strings.TrimSpace(*email)
    if len(*email) > 3 {
        if match, _ := regexp.MatchString("^.+@.+\\..+$", *email); match {
			if strings.Index(*email, " ") < 0 {  //no spaces
				return true
			}
			/*
			v1, _ := this.gun.ValidateEmail (email)
			return v1.IsValid
			*/
        }
    }
    return false
}