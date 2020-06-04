/*! \file twilio.go
* 	\brief Class for handling twilio

	Sending of text messages is queued through twilio, so as long as we don't get a "too many requests" we're all set to keep firing
	https://support.twilio.com/hc/en-us/articles/115002943027-Understanding-Twilio-Rate-Limits-and-Message-Queues
*/

package toolz

import (
	"github.com/pkg/errors"

	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"regexp"
	"context"
	"encoding/json"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const twilioMsgUrl = "2010-04-01/Accounts/%s/Messages.json"

var (
	ErrType_twilioInvalidNumber			= errors.New("Twilio to phone number is invalid")
	ErrType_twilioRateLimitExceeded		= errors.New("Twilio message que limit reached")
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type TwilioConfig_t struct {
	SID, Token string
}

type twilioResponse_t struct {
	Message string
	User    struct {
		Id  int64
	}
	Success bool
}

type Twilio_c struct {

}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Sends an sms/mms message to the target number
*/
func (this *Twilio_c) SMS (ctx context.Context, config *TwilioConfig_t, to, from, outMsg, outMediaUrl string) error {
	// setup our url param variables
	vals := url.Values{}
	vals.Set("To", to)
	vals.Set("From", from)
	vals.Set("Body", outMsg)

	if len(outMediaUrl) > 0 { vals.Set ("MediaUrl", outMediaUrl) } // see if we have a media url to include, converst the message to MMS

	// generate a url
	url := url.URL {
		Scheme: "https",
		Host: "api.twilio.com",
	}
	url.Path = fmt.Sprintf (twilioMsgUrl, config.SID)
	url.RawQuery = vals.Encode()

	// generate a new request with our info
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
	if err != nil { return errors.WithStack (err) }

	req.SetBasicAuth (config.SID, config.Token)
	req.Header.Set ("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return errors.WithStack (err) }

	defer resp.Body.Close()

	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return errors.WithStack (err) }

	if resp.StatusCode == 429 { // special error for a rate limit being exceeded, this message didn't send and we need to wait to try to send it again
		return errors.Wrap (ErrType_twilioRateLimitExceeded, from) // these are limited by the phone number, so track that with the error
	} else if resp.StatusCode > 299 { // generally bad
		// see why it's bad
		var m struct {
			Code 	int64
		}

		err = json.Unmarshal (rBody, &m)
		if err != nil { return errors.Wrapf (err, string(rBody)) }

		switch m.Code {
		case 21211:	//indicates the to number is not a valid phone number
			return errors.Wrapf (ErrType_twilioInvalidNumber, "Error sending twilio sms :: %d : %s", resp.StatusCode, to)
		default:
			//number should still be valid, something else is wrong
			return errors.Errorf ("Error sending twilio sms :: %d : %s : %s : %s", resp.StatusCode, to, from, string(rBody))
		}
	}
	return nil 	//we're good
}

/*! \brief Validates a phone number and returns the twilio approved format for it
*/
func (this *Twilio_c) ValidatePhoneNumber (original string) (out string, err error) {
	r := regexp.MustCompile("[^0-9]")
	out = r.ReplaceAllString(original, "")
	switch len(out) {
	case 10:
		out = "+1" + out
	case 11, 12, 13:
		out = "+" + out
	default:
		err = errors.Wrap (ErrType_twilioInvalidNumber, original)
		out = "" //clear this
	}
	return
}
