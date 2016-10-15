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
	var (
		s *database.SESSION
		p *database.PERSON
		k []*database.PUBLIC_KEY
	)
	confirmed := false
	alert := new(Alert)

	if "POST" == r.Method {
		r.ParseForm()

		sessionCode, sessionCodeExists := r.PostForm["sessionCode"]
		if sessionCodeExists {
			code := strings.Join(sessionCode, "")
			if len(code) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					session, sessionErr := ConfirmSessionCode(code, stmt[database.SESSION_CLEANUP], stmt[database.SESSION_LOOKUP_BY_CODE])
					if sessionErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if !session.Verified {
						session.Verified = true
						if session.Update(stmt[database.SESSION_UPDATE]) != nil {
							alert.AsError(OTHER_ERROR)
							return
						}
					}

					// attempt to find the person for this session
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], session.PersonId)
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

					if !person.Verified {
						person.Verified = true
						if person.Update(stmt[database.PERSON_UPDATE]) != nil {
							alert.AsError(OTHER_ERROR)
							return
						}
					}

					keys, keysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if keysErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(keys) == 0 {
						alert.AsError(NO_KEYS)
						return
					}

					// success
					s = session
					p = person
					k = keys
					confirmed = true
				}

				database.WithDatabase(db, fn)
			}
		}
	}

	if confirmed {
		postForm := &NewPostPage{Title: "New Post", Alert: alert, Session: s, Person: p, Keys: k}
		NEW_POST_TEMPLATE.Execute(w, postForm)
	} else {
		alert.Message = "If you did not get an email with a code to decrypt, you can <a href=\"/session\">request one here</a>"
		sessionForm := &ConfirmSessionPage{Title: "Confirm Session", Alert: alert}
		CONFIRM_SESSION_TEMPLATE.Execute(w, sessionForm)
	}
}
