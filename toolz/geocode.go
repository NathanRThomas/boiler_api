/*! \file geocode.go
 *  \brief Class for handling google's reverse geocoding, gets an address from long and lat
 */

package toolz

import (
    "fmt"
    "net/http"
    "io/ioutil"
    
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const googleGeocodeURL = "https://maps.googleapis.com/maps/api/geocode/json?latlng=%s,%s&key=%s"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type geocodeResponse_t struct {
    Results     []struct {
        Components  []struct {
            LongName    string  `json:"long_name"`
            Types       []string `json:"types"`
        }   `json:"address_components"`
    }   `json:"results"`
}

type Geocode_c struct { }

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Handles sending a json body to a url
 */
func (this *Geocode_c) send (lat, long string, res interface{}) {
    response, err := http.Get(fmt.Sprintf(googleGeocodeURL, lat, long, AppConfig.Google.Key))
    if err != nil {
        ErrChk(err)
    } else {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        //fmt.Println(string(contents[:]))
        if err != nil {
            ErrChk(err)
        } else {
            UM(contents, res)
        }
    }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Returns the state abbreviation based on the zip code
*/
func (this *Geocode_c) State (lat, long string) string {
    res := &geocodeResponse_t {}
    this.send (lat, long, res)
    //fmt.Printf("%+v\n", res)
    for _, addr := range res.Results {
        for _, c := range addr.Components {
            for _, ty := range c.Types {
                if ty == "administrative_area_level_1" { return c.LongName } //this is the level we're looking for
            }
        }
    }

    Err("Unalbe to find state: %s : %s", lat, long)
    return ""   //couldn't find the state
}

func (this *Geocode_c) Zip (lat, long string) string {
    res := &geocodeResponse_t {}
    this.send (lat, long, res)
    //fmt.Printf("%+v\n", res)
    for _, addr := range res.Results {
        for _, c := range addr.Components {
            for _, ty := range c.Types {
                if ty == "postal_code" { return c.LongName } //this is the level we're looking for
            }
        }
    }

    Err("Unalbe to find zip: %s : %s", lat, long)
    return ""   //couldn't find the state
}

