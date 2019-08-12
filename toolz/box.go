/*! \file box.go
 *  \brief Contains non-class based functions that can be re-used everywhere
 */

package toolz

import (
	"time"
	"crypto/sha256"
    "fmt"
	"os"
	"encoding/hex"
    "runtime"
	"strings"
	"strconv"
    "log"
	"encoding/json"
	"regexp"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFINES -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const (
    ProductionType_Dev      = iota
    ProductionType_Production
    ProductionType_Staging
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONFIG OBJECTS ----------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var AppConfig struct {
    ProductionFlag  int
    MachineName string
    Redis         	struct {
        IP         string
        Ports       []int
    }
    Cockroach       struct {
        IP, Database string
        Port        int
    }
    Cassandra       struct {
		Ips         []string
		Keyspace	string
    }
    Couch           struct {
        IP, Username, Password string
    }
    Plivo           struct {
        Number, AuthID, Token string
    }
    Stripe          struct {
        Publish, Secret string
    }
    Google          struct {
        Key string
	}
	DigitalOcean struct {
		Key string
	}
	Paypal 			struct {
        Client, Secret, API, Redirect, Cancel string
    }
	Twilio twilioConfig_t
	Slack slackConfig_t
	MailGun mailGunConfig_t
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func reverse (in []string) []string {
    if len(in) < 2 { return in }
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
        	in[i], in[j] = in[j], in[i]
	}
	return in
}

/*! \brief Loads our application configuration file.  Allows us to give personality to this application
 */
func LoadConfig() {
    configFile, err := os.Open(os.Getenv("API_CONFIG")) //try the file
    if err == nil {
        jsonParser := json.NewDecoder(configFile)
        err = jsonParser.Decode(&AppConfig)
    }
    if err != nil {
        log.Panicln(err)    //we can't move forward from here no matter what
    }
}

func StrToInt64(in string) int64 {
	id, _ := strconv.ParseInt(in, 10, 64)
	return id
}

/*! \brief When there is an error, this is used to figure out the calling stack for where it happened
 */
func Trace(err error) (error) {
    if err == nil { return nil }
    errList := make([]string, 0)
    
    for i := 0; i < 9; i++ {
        _, file, line, ok := runtime.Caller (i)
        if ok {
            fileParts := strings.Split(file, "/")
            for idx, p := range(fileParts) {
                if p == "go_api" {  //this is part of our code
                    errList = append(errList, fmt.Sprintf("%s/%s @ %d", fileParts[idx+1], fileParts[idx+2], line))
                }
            }
        } else {
            break
        }
    }
    
    errList[0] = err.Error()   //stomp the first error with the incoming one.  The last error is this function
    return fmt.Errorf("%s", strings.Join(errList, " :: "))
}

func ErrChk (err error) error {
    err = Trace(err)
    if err != nil { log.Println(err) }
    return err
}

func Err (msg string, params ...interface{}) error {
    return ErrChk(fmt.Errorf(msg, params...))
}

/*! \brief Tries to unmarshal an object and records an error if it happens
*/
func UM (jAttr []byte, out interface{}) error {
    err := ErrChk(json.Unmarshal(jAttr, out))
    if err != nil { //this didn't work
        if len(jAttr) >= 2 {
            Err("%s", string(jAttr[:])) 
        } else {
            Err("json input is empty")
        }
    }
    return err
}

func Reverse(input string) string {
    n := 0
    rune := make([]rune, len(input))
    for _, r := range input {   // Get Unicode code points. 
        rune[n] = r; n++
    } 
    rune = rune[0:n]
    // Reverse 
    for i := 0; i < n/2; i++ { 
        rune[i], rune[n-1-i] = rune[n-1-i], rune[i] 
    }
    return string(rune) // Convert back to UTF-8. 
}

/*! \brief Creates our simple hash based on a 256 byte hash, from the input string
 *  \return string that's the hash, ie a86049198f4168a44d352990717fdaa984909dec41838e1a5ba3e0973f777878
 */
func Hash(seed string, length int) string {
    if length <= 0 { length = 64 }	//give them all of it
    if len(seed) < 1 { seed = fmt.Sprintf("%d", time.Now().UnixNano()) }
    hash := sha256.Sum256([]byte(seed))
    return hex.EncodeToString(hash[:])[:length]
}

/*! \brief Validates the password meets our requirements
 */
 func ValidatePassword(pass string) error {
	matches := []string{"[a-z]", "[A-Z]"}
	if len(pass) < 8 {
		return fmt.Errorf("Password invalid! Must be at least 8 characters")
	}

	for _, m := range matches {
		r, _ := regexp.Compile(m)
		if len(r.FindAllString(pass, 1)) == 0 { //ensure we have at least one match
			return fmt.Errorf("Password invalid! Must contain at least 1 lower and 1 upper case letter")
		}
	}
	return nil //if we're here the password was good
}

/*! \brief Validates a phone number
 */
func ValidatePhone(phone *string) error {
	match := regexp.MustCompile(`^[\+1 -\(\)]*([2-9]\d{2})[\(\)\. -]{0,2}(\d{3})[\. -]?(\d{4})$`)
	resp := match.FindStringSubmatch(strings.Trim(*phone, " "))
	if len(resp) == 4 {
		*phone = fmt.Sprintf("1%s%s%s", resp[1], resp[2], resp[3]) //format this how we want it
	} else {
		return fmt.Errorf("Phone number appears invalid: %s", *phone)
	}
	return nil
}
