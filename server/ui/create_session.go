// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
	"strings"
)

type CreateSessionPage struct {
	Title string
	Alert *Alert
}

func CreateSession(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	alert := new(Alert)
	alert.Message = "If you do not have a public key associated with your email address, you can <a href=\"/upload\">upload it here</a>"

	if "POST" == r.Method {
		r.ParseForm()

		// an email address means create a new session, if there is at least one public key associated
		em, emExists := r.PostForm["userEmail"]
		if emExists {
			email := strings.ToLower(strings.Join(em, ""))
			if len(email) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					// attempt to find the person for this email address
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_EMAIL], email)
					if personErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(person.Id) == 0 {
						alert.AsError(UNKNOWN)
						return
					}

					if !person.Enabled {
						alert.AsError(DISABLED)
						return
					}

					// find this person's public keys
					publicKeys, publicKeysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if publicKeysErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(publicKeys) == 0 {
						alert.Update("alert-warning", "fa-hand-paper-o", NO_KEYS)
						return
					}

					sessionErr := CreateNewSession(person, publicKeys, stmt[database.SESSION_INSERT])
					if sessionErr != nil {
						alert.AsError(sessionErr.Error())
						return
					} else {
						// present the session code form
						Redirect("/confirm")(w, r)
					}
				}

				database.WithDatabase(db, fn)
			}
		}
	}

	sessionForm := &CreateSessionPage{Title: TITLE_CREATE_SESSION, Alert: alert}
	CREATE_SESSION_TEMPLATE.Execute(w, sessionForm)
}
