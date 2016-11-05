// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"bytes"
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/cryptutil"
	"github.com/Banrai/TeamWork.io/server/database"
	"io"
	"net/http"
	"strings"
)

type NewKeyPage struct {
	Title   string
	Alert   *Alert
	Session *database.SESSION
	Person  *database.PERSON
}

func UploadKey(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	alert := new(Alert)

	if "POST" == r.Method {
		r.ParseMultipartForm(16384)

		fn := func(stmt map[string]*sql.Stmt) {
			// attempt to read the uploaded public key file
			pkFile, pkFileHeader, pkFileErr := r.FormFile("publicKey")
			if pkFileErr != nil {
				alert.AsError(INVALID_PK)
				return
			}
			defer pkFile.Close()

			buf := new(bytes.Buffer)
			_, copyErr := io.Copy(buf, pkFile)
			if copyErr != nil {
				alert.AsError(INVALID_PK)
				return
			}

			// make sure the uploaded public key is valid
			uploadedKey := buf.String()
			_, invalidKeyErr := cryptutil.DecodeArmoredKey(uploadedKey)
			if invalidKeyErr != nil {
				alert.AsError(INVALID_PK)
				return
			}

			// an email address posted in the form request
			em, emExists := r.PostForm["userEmail"]
			if emExists {
				// an email address should have been provided
				email := strings.ToLower(strings.Join(em, ""))
				if len(email) == 0 {
					alert.AsError(NO_EMAIL)
					return
				} else {
					// attempt to find the person for this email address
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_EMAIL], email)
					if personErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(person.Id) == 0 {
						// this is a new person
						person.Email = email
						personId, personAddErr := person.Add(stmt[database.PERSON_INSERT])
						if personAddErr != nil {
							alert.AsError(OTHER_ERROR)
							return
						}
						person.Id = personId
					}

					// now add this key to the database for this person
					pkErr := AddPublicKey(person, uploadedKey, KEY_SOURCE, pkFileHeader.Filename, stmt[database.PK_INSERT])
					if pkErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					// find all this person's public keys
					publicKeys, publicKeysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if publicKeysErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					// create the session, and ask for confirmation of the decrypted code
					sessionErr := CreateNewSession(person, publicKeys, stmt[database.SESSION_INSERT])
					if sessionErr != nil {
						alert.AsError(sessionErr.Error())
						return
					} else {
						// present the session code form
						Redirect("/confirm")(w, r)
					}
				}
			}

			// versus a pre-existing session
			sessionCode, sessionCodeExists := r.PostForm["session"]
			personCode, personCodeExists := r.PostForm["person"]
			if sessionCodeExists && personCodeExists {
				sessionId := strings.Join(sessionCode, "")
				personId := strings.Join(personCode, "")
				if len(sessionId) > 0 && len(personId) > 0 {
					// make sure the session is still valid
					session, sessionErr := ConfirmSessionCode(sessionId, stmt[database.SESSION_CLEANUP], stmt[database.SESSION_LOOKUP_BY_CODE])
					if sessionErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if !session.Verified {
						alert.AsError(INVALID_SESSION)
						return
					}

					// attempt to find the person for this session
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], session.PersonId)
					if personErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(person.Id) == 0 || !person.Verified {
						alert.AsError(UNKNOWN)
						return
					}

					if !person.Enabled {
						alert.AsError(DISABLED)
						return
					}

					// now add this key to the database for this person
					pkErr := AddPublicKey(person, uploadedKey, KEY_SOURCE, pkFileHeader.Filename, stmt[database.PK_INSERT])
					if pkErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					// find all this person's public keys
					keys, keysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if keysErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					// create the session, and ask for confirmation of the decrypted code
					sessionCreateErr := CreateNewSession(person, keys, stmt[database.SESSION_INSERT])
					if sessionCreateErr != nil {
						alert.AsError(sessionCreateErr.Error())
						return
					} else {
						// present the session code form
						Redirect("/confirm")(w, r)
					}
				}
			}

			// session information is missing, or the request is otherwise invalid,
			// so prompt with a new session form request
			Redirect("/session")(w, r)
		}

		database.WithDatabase(db, fn)
	}

	page := &NewKeyPage{Title: TITLE_ADD_KEY, Alert: alert, Session: new(database.SESSION), Person: new(database.PERSON)}
	NEW_KEY_TEMPLATE.Execute(w, page)
}
