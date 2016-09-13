// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
	"strings"
)

type ConfirmSessionPage struct {
	Title string
	Alert *Alert
}

func ConfirmSession(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	alert := new(Alert)
	alert.Message = "If you did not get an email with a code to decrypt, you can <a href=\"/session\">request one here</a>"

	if "POST" == r.Method {
		r.ParseForm()

		sessionCode, sessionCodeExists := r.PostForm["sessionCode"]
		if sessionCodeExists {
			code := strings.Join(sessionCode, "")
			if len(code) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					// remove any expired sessions
					database.CleanupSessions(stmt[database.SESSION_CLEANUP])

					// fetch the session corresponding to this code
					session, sessionErr := database.LookupSession(stmt[database.SESSION_LOOKUP_BY_CODE], code)
					if sessionErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
					}

					if !session.Verified {
						session.Verified = true
						if session.Update(stmt[database.SESSION_UPDATE]) != nil {
							alert.AlertType = "alert-danger"
							alert.Icon = "fa-exclamation-triangle"
							alert.Message = OTHER_ERROR
							return
						}
					}

					// attempt to find the person for this session
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], session.PersonId)
					if personErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
					}

					if !person.Enabled {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = DISABLED
						return
					}

					if !person.Verified {
						person.Verified = true
						if person.Update(stmt[database.PERSON_UPDATE]) != nil {
							alert.AlertType = "alert-danger"
							alert.Icon = "fa-exclamation-triangle"
							alert.Message = OTHER_ERROR
							return
						}
					}

					// success: present the new message form
					postForm := &NewPostPage{Title: "New Post", Session: session, Person: person}
					NEW_POST_TEMPLATE.Execute(w, postForm)
				}

				database.WithDatabase(db, fn)
			}
		}
	}

	sessionForm := &ConfirmSessionPage{Title: "Confirm Session", Alert: alert}
	CONFIRM_SESSION_TEMPLATE.Execute(w, sessionForm)
}