/*! \file twilio.go
 *  \brief Class for handling 2 factor authentication
 */

package toolz

import (
    "fmt"
    "encoding/json"
    "net/http"
    "net/url"
    //"bytes"
    "io/ioutil"
    
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const twilio_base_url    = "https://api.authy.com/protected/json/"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type twilioConfig_t struct {
	SID, Token string
}

type twilioResponse_t struct {
    Msg     string  `json:"message"`
    User    struct {
        ID  int64   `json:"id"`
    } `json:"user"`
    Success bool    `json:"success"`
}

type twilioVeiryResponse_t struct {
    Msg     string  `json:"message"`
    Token   string  `json:"token"`
    Success string  `json:"success"`
}

type Twilio_c struct {
    
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Handles sending a json body to a url
 */
func (this *Twilio_c) send (url, method string, vals url.Values, tResp interface{}) {
    //if AppConfig.ProductionFlag != ProductionType_Production { return false } //only send twilio messages on production

    req, err := http.NewRequest(method, twilio_base_url + url + "?" + vals.Encode(), nil)
    if err != nil {
        ErrChk(fmt.Errorf("twilio genUrl Failed: " + err.Error()))
        return
    }
    
    //req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("X-Authy-API-Key", AppConfig.Twilio.Token)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        ErrChk(fmt.Errorf("twilio client Failed: " + err.Error()))
        return
    }
    defer resp.Body.Close()

    //fmt.Printf("%+v\n", resp)
    rBody, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode > 299 {
        //fmt.Println("response Status:", resp.Status)
        //fmt.Println("response Headers:", resp.Header)
        err = ErrChk(fmt.Errorf("twilio Client Failed: %d : %s : %s", resp.StatusCode, url, string(rBody[:])))
    }

    ErrChk(json.Unmarshal(rBody, &tResp))
    return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Creates a new user with twilio auty
*/
func (this *Twilio_c) Create (email, fullPhone string) int64 {
    code := "1"
    phone := fullPhone

    if len(phone) == 11 {   //means the first digit is the country code
        code = fullPhone[0:1]
        phone = fullPhone[1:]
    } else if len(fullPhone) < 10 {
        ErrChk(fmt.Errorf("Phone number is invalid : %s : %s", email, fullPhone))
        return 0
    }

    v := url.Values{}
    v.Set("user[email]", email)
    v.Set("user[cellphone]", phone)
    v.Set("user[country_code]", code)

    resp := &twilioResponse_t{}
    this.send ("users/new", "POST", v, resp)
    return resp.User.ID     //return the id of interest
}

/*! \brief Sends a token to the authy user for them to send back to us to verify
*/
func (this *Twilio_c) SendToken (authyID int64) bool {
    resp := &twilioResponse_t{}
    this.send (fmt.Sprintf("sms/%d", authyID), "GET", url.Values{}, resp)
    return resp.Success     //return see if it worked
}

/*! \brief Verifies a token against a user's account
*/
func (this *Twilio_c) Verify (token string, authyID int64) bool {
    resp := &twilioVeiryResponse_t{}
    this.send (fmt.Sprintf("verify/%s/%d", token, authyID), "GET", url.Values{}, resp)
    return resp.Success == "true" && resp.Token == "is valid"     //return see if it worked
}