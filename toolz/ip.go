/*! \file ip.go
 *  \brief Class for handling ip addresses
 *  For determining if this is a facebook request or not
 */

package toolz

import (
    //"fmt"
    "net"
    "net/http"
    "bytes"
    "strings"
    "regexp"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type ipRange struct {
    start net.IP
    end net.IP
}

var privateRanges = []ipRange{
    ipRange{
        start: net.ParseIP("10.0.0.0"),
        end:   net.ParseIP("10.255.255.255"),
    },
    ipRange{
        start: net.ParseIP("100.64.0.0"),
        end:   net.ParseIP("100.127.255.255"),
    },
    ipRange{
        start: net.ParseIP("172.16.0.0"),
        end:   net.ParseIP("172.31.255.255"),
    },
    ipRange{
        start: net.ParseIP("192.0.0.0"),
        end:   net.ParseIP("192.0.0.255"),
    },
    ipRange{
        start: net.ParseIP("192.168.0.0"),
        end:   net.ParseIP("192.168.255.255"),
    },
    ipRange{
        start: net.ParseIP("198.18.0.0"),
        end:   net.ParseIP("198.19.255.255"),
    },
}

type IP_c struct {
    tutor Tutor
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *IP_c) inRange(r ipRange, ipAddress net.IP) bool {
    // strcmp type byte comparison
    if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
        return true
    }
    return false
}

func (this *IP_c) isPrivateSubnet(ipAddress net.IP) bool {
    // my use case is only concerned with ipv4 atm
    if ipCheck := ipAddress.To4(); ipCheck != nil {
        // iterate over all our ranges
        for _, r := range privateRanges {
            // check if this ip is in a private range
            if this.inRange(r, ipAddress){
                return true
            }
        }
    }
    return false
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *IP_c) GetIPAddress(r *http.Request) string {
    for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
        for _, ip := range strings.Split(r.Header.Get(h), ",") {
            // header can contain spaces too, strip those out.
            ip = strings.TrimSpace(ip)
            realIP := net.ParseIP(ip)
            if !realIP.IsGlobalUnicast() || this.isPrivateSubnet(realIP) {
                // bad address, go to next
                continue
            } else {
                return ip
            }
        }
    }
    return ""   //couldn't find it
}

func (this *IP_c) IsPhone (str string) (match bool) {
    match, _ = regexp.MatchString(`^\+1\d{10}$`, str)
    return
}

/*! \brief Handles the regex for parsing out what the url was
 */
func (this *IP_c) ParseUrlParts(path, regex string) (parts []string) {
	if len(regex) == 0 { regex = `^\/(\w+)\/(\w+)` }
	match, err := regexp.Compile(regex)
	if err == nil {
		resp := match.FindStringSubmatch(path) //parse our url to figure out where we're going
		for i, p := range resp {
			if i > 0 { 
                parts = append(parts, strings.ToLower(p))
            }
		}
	} else { Err("%s :: %s", err.Error(), path) }
	return
}

/*! \brief This handles breaking our url into what we're expecting for our own internal redirecting
*/
func (this *IP_c) ParseLinksUrl (url, agent string) (parts []string, ok bool) {
	match := regexp.MustCompile(`^\/(\w+)\/([\w-]+)\/?([\w-]+)?`)
	resp := match.FindStringSubmatch(url)
	
	if len(resp) > 1 {	//we got what we were expecting
		parts = resp[1:]	//ignore the first one cause it's just the url again
		ok = true 
	}
	return
}
