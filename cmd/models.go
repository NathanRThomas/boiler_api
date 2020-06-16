/*! \file models.go
	\brief Defines needed at the cmd level
*/

package cmd 

 import (
	"github.com/NathanRThomas/boiler_api/pkg/models"
	
	"github.com/patrickmn/go-cache"
	
 )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const ContextTimeout		= 50	// number of seconds a single task/context should be allowed to run, we use this with shutting down as well

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TYPES -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const (
	ApiErrorCode_internal					= iota + 1
	ApiErrorCode_passwordGuessing
	ApiErrorCode_parsingRequestBody
	ApiErrorCode_noIdentifiersForUser
	
	ApiErrorCode_notEnoughInfoToCreateUser // 5
	ApiErrorCode_invalidInputField
	ApiErrorCode_emailExistsAlready
	ApiErrorCode_endpointDoesNotExist
	ApiErrorCode_jsonMarshal

	ApiErrorCode_panicRecovery  	// 10
	ApiErrorCode_missingUser
	ApiErrorCode_dbError
	ApiErrorCode_invalidUrlParam
	ApiErrorCode_thirdPartyRequest

	ApiErrorCode_missingFromContext 	// 15
	ApiErrorCode_range

) 

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

//----- SIGNUP -----//
type SignupUser_t struct {
	Email, Password, Phone models.ApiString
	
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CACHE FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Wrapper around getting a value from the database and saving it in local cache
*/
func (this *App_c) GetUser (userID models.UUID) (*models.User_t, error) {
	key := userID.Key ("user")
	if data, found := this.Cache.Get (key); found { return data.(*models.User_t), nil }	// we're cached

	user := &models.User_t { ID: userID }
	err := this.Users.Get (user)
	if err != nil { return nil, err }

	this.Cache.Set (key, user, cache.DefaultExpiration) // now cache it for next time
	return user, nil
}

