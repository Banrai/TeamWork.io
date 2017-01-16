// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/api"
	"github.com/Banrai/TeamWork.io/server/cryptutil"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/emailer"
	"github.com/Banrai/TeamWork.io/server/httputil"
	"html/template"
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
	var (
		s *database.SESSION
		p *database.PERSON
	)
	alert := new(Alert)
	alert.Message = "Please use a public key (in ASCII-armored format) which corresponds to this email"

	if "POST" == r.Method {
		r.ParseMultipartForm(16384)

		fn := func(stmt map[string]*sql.Stmt) {
			// see if this is an in-session request
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

					// session and person are established
					s = session
					p = person
				}
			}

			// an email address should have been provided
			em, emExists := r.PostForm["userEmail"]
			if !emExists {
				alert.AsError(NO_EMAIL)
				return
			}

			// check its validity
			email := strings.ToLower(strings.Join(em, ""))
			if !emailer.IsPossibleEmail(email) {
				alert.AsError(INVALID_EMAIL)
				return
			}

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

			// find all this person's public keys
			publicKeys, publicKeysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
			if publicKeysErr != nil {
				alert.AsError(OTHER_ERROR)
				return
			}

			// determine the public key source: file or url
			kt, ktExists := r.PostForm["keyType"]
			if !ktExists {
				alert.AsError(api.INVALID_REQUEST)
				return
			}

			keyType := strings.Join(kt, "")
			if "upload" == keyType {
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

				// find out if this key already exists
				alreadyExists := false
				for _, priorKey := range publicKeys {
					if uploadedKey == priorKey.Key {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					// now add this key to the database for this person
					pkErr := AddPublicKey(person, uploadedKey, KEY_SOURCE, pkFileHeader.Filename, stmt[database.PK_INSERT])
					if pkErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}
				}
			} else {
				// source is a url
				u, uExists := r.PostForm["publicKeyUrl"]
				if !uExists {
					alert.AsError(api.INVALID_REQUEST)
					return
				}

				url := strings.Join(u, "")
				urlKey, urlKeyErr := httputil.URLFetchAsString(url)
				if urlKeyErr != nil {
					alert.AsError(urlKeyErr.Error())
					return
				}

				// make sure the fetched url resource is valid
				_, invalidKeyErr := cryptutil.DecodeArmoredKey(urlKey)
				if invalidKeyErr != nil {
					alert.AsError(INVALID_PK)
					return
				}

				// find out if this key already exists
				alreadyExists := false
				for _, priorKey := range publicKeys {
					if urlKey == priorKey.Key {
						alreadyExists = true
						break
					}
				}

				if !alreadyExists {
					// now add this key to the database for this person
					pkErr := AddPublicKey(person, urlKey, KEY_SOURCE, url, stmt[database.PK_INSERT])
					if pkErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}
				}
			}

			_, createSessionExists := r.PostForm["createSession"]
			if createSessionExists {
				// create the session, and ask for confirmation of the decrypted code
				sessionErr := CreateNewSession(person, publicKeys, stmt[database.SESSION_INSERT])
				if sessionErr != nil {
					alert.AsError(sessionErr.Error())
					return
				} else {
					// present the session code form
					Redirect("/confirm")(w, r)
				}
			} else {
				alert.Message = template.HTML(fmt.Sprintf("The public key for \"%s\" was added successfully", email))
			}
		}

		database.WithDatabase(db, fn)
	}

	if s == nil && p == nil {
		s = new(database.SESSION)
		p = new(database.PERSON)
	}

	page := &NewKeyPage{Title: TITLE_ADD_KEY, Alert: alert, Session: s, Person: p}
	NEW_KEY_TEMPLATE.Execute(w, page)
}
