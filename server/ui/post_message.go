// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
	"strings"
)

type NewPostPage struct {
	Title   string
	Alert   *Alert
	Session *database.SESSION
	Person  *database.PERSON
	Keys    []*database.PUBLIC_KEY
}

func PostMessage(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	alert := new(Alert)

	if "POST" == r.Method {
		r.ParseForm()

		sessionCode, sessionCodeExists := r.PostForm["session"]
		if sessionCodeExists {
			sessionId := strings.Join(sessionCode, "")
			if len(sessionId) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					// remove any expired sessions
					database.CleanupSessions(stmt[database.SESSION_CLEANUP])

					// fetch the session corresponding to this id
					session, sessionErr := database.LookupSession(stmt[database.SESSION_LOOKUP_BY_ID], sessionId)
					if sessionErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
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
				}
				database.WithDatabase(db, fn)
			}
		}
	}

	postForm := &NewPostPage{Title: "New Post", Session: &database.SESSION{}, Person: &database.PERSON{}}
	NEW_POST_TEMPLATE.Execute(w, postForm)
}
