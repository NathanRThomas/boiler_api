/*! \file slack.go
 *  \brief Class for handling messaging channels using slack
 */

package toolz

import (
    "fmt"
    "strconv"
    "encoding/json"
    "net/http"
    "bytes"
    "io/ioutil"
    
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const slack_base_url    = "https://slack.com/api/"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type slackConfig_t struct {
	Username, Token string
}

type slack_body_t struct {
    Channel         string  `json:"channel,omitempty"`
    Text            string  `json:"text,omitempty"`
    Username        string  `json:"username,omitempty"`
}

type Slack_c struct {
    
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Handles sending a json body to a url
 */
func (this *Slack_c) send (url, requestType string, body slack_body_t) (bool) {
    if AppConfig.ProductionFlag != ProductionType_Production { return false } //only send slack messages on production

    body.Username = AppConfig.Slack.Username
    jsonBody, _ := json.Marshal(body)

    req, err := http.NewRequest(requestType, slack_base_url + url, bytes.NewBuffer(jsonBody))
    if err != nil {
        ErrChk(fmt.Errorf("FBM genUrl Failed: " + err.Error()))
        return false
    }
    
    //req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Content-Length", strconv.Itoa(len(jsonBody)))
    req.Header.Set("Authorization", "Bearer " + AppConfig.Slack.Token)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        ErrChk(fmt.Errorf("slack client Failed: " + err.Error()))
        return false
    }
    defer resp.Body.Close()

    //fmt.Printf("%+v\n", resp)
    rBody, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode > 299 {
        //fmt.Println("response Status:", resp.Status)
        //fmt.Println("response Headers:", resp.Header)
        err = ErrChk(fmt.Errorf("slack Client Failed: %d : %s : %s : %s", resp.StatusCode, url, string(jsonBody[:]), string(rBody[:])))
    }

    var resJson struct {
        OK      bool    `json:"ok"`
        Error   string  `json:"error"`
    }
    ErrChk(json.Unmarshal(rBody, &resJson))

    if err != nil {
        err = ErrChk(fmt.Errorf("slack response failed : %s : %s", resJson.Error, err.Error()))
    } else if !resJson.OK {
        err = ErrChk(fmt.Errorf("slack response failed : %s", resJson.Error))
    }

    if err == nil {
        return true
    }
    return false
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *Slack_c) Channel (msg, channelID string) {
    this.send ("chat.postMessage", "POST", slack_body_t { Channel: channelID, Text: msg })
}

