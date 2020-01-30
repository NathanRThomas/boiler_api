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

/*! \brief Handles the actual insertion of a new user
*/
func (this *User_c) Insert (tx *sql.Tx, user *models.User_t) error {
	jAttr, err := json.Marshal(user.Attr)
	if err != nil { return errors.Wrapf (err, "%+v\n", user.Attr) }

	err = tx.QueryRow (`INSERT INTO users (attrs, mask) VALUES ($1, $2) RETURNING id`,
					jAttr, user.Mask).Scan(&user.ID)
	return errors.Wrapf (err, "%+v\n", user)
}

/*! \brief Gets our user from the database
*/
func (this *User_c) Get (tx *sql.Tx, user *models.User_t) error {
	if !user.ID.Valid() { return errors.WithStack (models.ErrType_invalidUUID) }  //this isn't good, can't find a user with the id this way

	var jAttr []byte
	err := tx.QueryRow(`SELECT mask, token, attrs, created FROM users WHERE id = $1`, 
			user.ID).Scan(&user.Mask, &jAttr, &user.Created)

	if err != nil { return errors.Wrap (err, user.ID.String()) }
	
	return this.UM(jAttr, &user.Attr)
}
