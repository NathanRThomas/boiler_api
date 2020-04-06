/*! \file user.go
	\brief Cockroach specific to user calls

*/

package cockroach

import (
	"github.com/NathanRThomas/boiler_api/pkg/models"

	"github.com/pkg/errors"

	//"fmt"
	"encoding/json"
	"database/sql"
)

type User_c struct {
	toolz_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Verifies that this email is new or returns the exsting info about that user
*/
func (this *User_c) FromEmail (email models.ApiString, userID models.UUID) (*models.User_t, error) {
	if !email.Email() { return nil, errors.WithStack (models.ErrType_noIdentifiers) }
	if !userID.Valid() { userID.Set("00000000-0000-0000-0000-000000000000") }	// otherwise we get an error about it not being a uuid in the query

	user := &models.User_t {}

    err := db.QueryRow(`SELECT id FROM users WHERE lower(email) = lower($1) AND mask & $3 = 0 AND id <> $2`, 
					email.String(), userID, models.UserMask_deleted).Scan(&user.ID)
	if err != nil { return nil, errors.Wrapf (err, "userID: %s :: email: %s", userID, email) }

	err = this.Get (user)
    return user, err
}

/*! \brief Creates a new user or updates an existing
*/
func (this *User_c) Save (user *models.User_t) error {
	if !user.Email.Email() { return errors.Wrap (models.ErrType_returnToUser, "Email appears invalid") }

	// verify it's a unique email
	existing, err := this.FromEmail (user.Email, user.ID)
	if existing != nil { return errors.Wrap (models.ErrType_returnToUser, "Email already in use by someone else") }

	switch errors.Cause (err) {
	case models.ErrType_noIdentifiers, sql.ErrNoRows, nil: // these are all fine

	default:
		return err // this is a bad one
	}

	jAttr, err := json.Marshal (user.Attr)
	if err != nil { return errors.WithStack (err) }

	if user.ID.Valid() { // we're updating
		err = this.Exec (`UPDATE users SET email = $1, attrs = $2 WHERE id = $3`, user.Email, jAttr, user.ID)
		if err != nil { return err }

		if user.Password.Valid() { // they don't have to set a password for updates
			user.SetToken()
			err = this.Exec(`UPDATE users SET password = $1, token = $2 WHERE id = $3`, user.Password.Hash(), user.Token, user.ID)
			if err != nil { return err }
		}
	} else { // we're inserting
		user.SetToken()
		err = db.QueryRow (`INSERT INTO users (email, password, token, attrs, mask)
							VALUES ($1, $2, $3, $4, $5) RETURNING id`, user.Email, 
							user.Password.Hash(), user.Token, jAttr, user.Mask).Scan(&user.ID)

		if err != nil { return errors.WithStack (err) }
	}
	return nil // we're good
}

/*! \brief Gets our user from the database
*/
func (this *User_c) Get (user *models.User_t) error {
	if !user.ID.Valid() { return errors.WithStack (models.ErrType_invalidUUID) }  //this isn't good, can't find a user with the id this way

	var jAttr []byte
	err := db.QueryRow(`SELECT mask, token, attrs, created FROM users WHERE id = $1`, 
			user.ID).Scan(&user.Mask, &jAttr, &user.Created)

	if err != nil { return errors.Wrap (err, user.ID.String()) }
	
	return this.UM(jAttr, &user.Attr)
}

/*! \brief Pulls the user based on the combo of their id and token
 */
 func (this *User_c) TokenLogin (user *models.User_t) error {
    //make sure we have good data
	if !user.ID.Valid() || !user.Token.Valid() { return errors.WithStack (models.ErrType_noIdentifiers) }
	
	var id models.UUID

	err := db.QueryRow(`SELECT id FROM users WHERE id = $1 AND token = $2 AND mask & $3 = 0`,
						user.ID, user.Token, models.UserMask_deleted).Scan(&id)
	if err != nil { return errors.WithStack (err) }

	if id != user.ID { return errors.WithStack (sql.ErrNoRows) } // this didn't work
	return this.Get (user) // finish our get
}

/*! \brief Default logging in
*/
func (this *User_c) Login (user *models.User_t) error {
	if user.Email.Email() && user.Password.Valid() {
		err := db.QueryRow(`SELECT id FROM users WHERE password = $2 AND lower(email) = lower($1) AND mask & $3 = 0`,
							user.Email, user.Password.Hash(), models.UserMask_deleted).Scan(&user.ID)
		if err != nil { return errors.WithStack (err) }
	}

	return this.Get (user)
}
