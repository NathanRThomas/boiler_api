/*! \file slack.go
 *  \brief Class for communicating with slack messages
 */

package toolz

import (

	"github.com/pkg/errors"

    "net/http"
    "io/ioutil"
	"context"
	"encoding/json"
	"strconv"
	"bytes"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const slack_base_url    = "https://slack.com/api/"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type SlackConfig_t struct {
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
func (this *Slack_c) send (ctx context.Context, url, requestType string, body slack_body_t) error {
	//if AppConfig.ProductionFlag != ProductionLevel_Production { return false } //only send slack messages on production
	config, ok := ctx.Value ("slackConfig").(*SlackConfig_t) // get our config
	if !ok { return errors.New ("slack config missing from context") }

    body.Username = config.Username
	jsonBody, err := json.Marshal(body)
	if err != nil { return errors.WithStack (err) }

    req, err := http.NewRequestWithContext(ctx, requestType, slack_base_url + url, bytes.NewBuffer(jsonBody))
    if err != nil { return errors.Wrap (err, "FBM genUrl Failed") }
    
    //req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Content-Length", strconv.Itoa(len(jsonBody)))
    req.Header.Set("Authorization", "Bearer " + config.Token)
    
    resp, err := http.DefaultClient.Do(req)
	if err != nil { return errors.Wrap (err, "slack client Failed") }
    defer resp.Body.Close()

    //fmt.Printf("%+v\n", resp)
    rBody, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode > 299 {
        //fmt.Println("response Status:", resp.Status)
        //fmt.Println("response Headers:", resp.Header)
		return errors.Errorf("slack Client Failed: %d : %s : %s", resp.StatusCode, url, string(rBody[:]))
    }

    var resJson struct {
        OK      bool    `json:"ok"`
        Error   string  `json:"error"`
    }
	err = json.Unmarshal(rBody, &resJson)
	if err != nil { return errors.Wrap (err, string(rBody)) }

	if !resJson.OK { return errors.Errorf ("slack response failed: %s", resJson.Error) }
	return nil // we're good
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *Slack_c) Error (ctx context.Context, msg string) error {
    return this.send (ctx, "chat.postMessage", "POST", slack_body_t { Channel: "GK5GV8GG2", Text: msg })
}
