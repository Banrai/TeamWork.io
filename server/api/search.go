// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/keyservers"
	"log"
	"net/http"
	"strings"
)

// Respond to an ajax request: return all the public keys for this person from the database
func GetPersonPublicKeys(r *http.Request, db database.DBConnection) string {
	// the result is a json representation of the list of public keys found
	results := make([]*database.PUBLIC_KEY, 0)
	valid := false

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		s, sExists := r.PostForm["sessionId"]
		if !sExists {
			return GenerateSimpleMessage(INVALID_REQUEST, MISSING_PARAMETER)
		}
		sessionId := strings.Join(s, "")

		p, pExists := r.PostForm["personId"]
		if !pExists {
			return GenerateSimpleMessage(INVALID_REQUEST, MISSING_PARAMETER)
		}

		fn := func(stmt map[string]*sql.Stmt) {
			// remove any expired sessions
			database.CleanupSessions(stmt[database.SESSION_CLEANUP])

			// find the person
			person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], strings.Join(p, ""))
			if personErr != nil {
				return
			}

			// find all of its associated sessions
			personSessions, personSessionErr := person.LookupSessions(stmt[database.SESSION_LOOKUP_BY_PERSON])
			if personSessionErr != nil {
				return
			}

			// are any of them valid?
			for _, session := range personSessions {
				if session.Id == sessionId {
					if session.Verified {
						valid = true
						break
					}
				}
			}

			if valid {
				// get all the associated public keys
				personKeys, personKeysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
				if personKeysErr != nil {
					return
				}

				for _, pk := range personKeys {
					results = append(results, pk)
				}
			}
		}

		database.WithDatabase(db, fn)
	}

	if !valid {
		return GenerateSimpleMessage(INVALID_REQUEST, INVALID_SESSION)
	} else {
		result, err := json.Marshal(results)
		if err != nil {
			return GenerateSimpleMessage(INVALID_REQUEST, err.Error())
		}
		return string(result)
	}
}

// Respond to an ajax request: return all the public keys for this email,
// on behalf of the particular registered person, with a valid session
func SearchPersonPublicKeys(r *http.Request, db database.DBConnection) string {
	// the result is a json representation of the list of public keys found
	results := make([]*database.PUBLIC_KEY, 0)
	valid := false

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		s, sExists := r.PostForm["sessionId"]
		if !sExists {
			return GenerateSimpleMessage(INVALID_REQUEST, MISSING_PARAMETER)
		}
		sessionId := strings.Join(s, "")

		p, pExists := r.PostForm["personId"]
		if !pExists {
			return GenerateSimpleMessage(INVALID_REQUEST, MISSING_PARAMETER)
		}

		// the email address is the search parameter
		em, emExists := r.PostForm["email"]
		if !emExists {
			return GenerateSimpleMessage(INVALID_REQUEST, MISSING_PARAMETER)
		}

		fn := func(stmt map[string]*sql.Stmt) {
			// remove any expired sessions
			database.CleanupSessions(stmt[database.SESSION_CLEANUP])

			// find the person making the request
			person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], strings.Join(p, ""))
			if personErr != nil {
				return
			}

			// find all of its associated sessions
			personSessions, personSessionErr := person.LookupSessions(stmt[database.SESSION_LOOKUP_BY_PERSON])
			if personSessionErr != nil {
				return
			}

			// are any of them valid?
			for _, session := range personSessions {
				if session.Id == sessionId {
					if session.Verified {
						valid = true
						break
					}
				}
			}

			if valid {
				searchEmail := strings.ToLower(strings.Join(em, ""))
				// see if there any public keys for the given email address already in the db,
				// based on existing person registrations
				searchPerson, searchPersonErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_EMAIL], searchEmail)
				if len(searchPerson.Id) == 0 || searchPersonErr != nil {
					// person with this email is currently unknown
					// see if the pk + email exist in the MIT key server
					keys, keysErr := keyservers.MITSearch(searchEmail)
					if keysErr != nil {
						return
					}

					for _, key := range keys {
						result := new(database.PUBLIC_KEY)
						result.Key = key
						result.Source = keyservers.MIT_SOURCE

						results = append(results, result)
					}

					// add the PERSON and each corresponding PUBLIC_KEY to the database w/o blocking
					go func() {
						if len(results) > 0 {
							log.Println(fmt.Sprintf("AddPersonWithKeys(): %s", searchEmail))
							err := database.AddPersonWithKeys(stmt[database.PERSON_INSERT], stmt[database.PK_INSERT], searchEmail, results)
							if err != nil {
								log.Println(err)
							}
						}
					}()
				} else {
					// email corresponds to an existing person in the db
					personKeys, personKeysErr := searchPerson.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if personKeysErr != nil {
						return
					}

					for _, pk := range personKeys {
						results = append(results, pk)
					}
				}
			}
		}

		database.WithDatabase(db, fn)
	}

	if !valid {
		return GenerateSimpleMessage(INVALID_REQUEST, INVALID_SESSION)
	} else {
		result, err := json.Marshal(results)
		if err != nil {
			return GenerateSimpleMessage(INVALID_REQUEST, err.Error())
		}
		return string(result)
	}
}
