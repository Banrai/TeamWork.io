// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
	"strings"
)

type DisplayPostsPage struct {
	Title   string
	Alert   *Alert
	Session *database.SESSION
	Person  *database.PERSON
	Posts   []*database.MESSAGE_DIGEST
}

func DisplayPosts(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	var (
		s *database.SESSION
		p *database.PERSON
		m []*database.MESSAGE_DIGEST
	)
	alert := new(Alert)
	confirmSession := false

	if "POST" == r.Method {
		r.ParseForm()
		confirmSession = true

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

					if len(session.Id) == 0 {
						alert.AsError(INVALID_SESSION)
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

					// found a valid session and person
					s = session
					p = person

					messages, _ := person.LookupLatestMessages(stmt[database.LATEST_MESSAGES], POSTS_PER_PAGE, 0)
					digests, _ := database.GetMessageDigests(stmt[database.PERSON_LOOKUP_BY_ID], stmt[database.RECIPIENTS_BY_MESSAGE], messages)
					m = digests
				}
				database.WithDatabase(db, fn)
			}
		}

	} else {
		// retrieve the latest digests, without session/person
		fn := func(stmt map[string]*sql.Stmt) {
			messages, _ := database.RetrieveMessages(stmt[database.LATEST_MESSAGES], "", POSTS_PER_PAGE, 0)
			digests, _ := database.GetMessageDigests(stmt[database.PERSON_LOOKUP_BY_ID], stmt[database.RECIPIENTS_BY_MESSAGE], messages)
			m = digests
		}
		database.WithDatabase(db, fn)

		// define these as empty, so the session template renders properly
		s = new(database.SESSION)
		p = new(database.PERSON)

	}

	if confirmSession && s == nil && p == nil {
		sessionForm := &ConfirmSessionPage{Title: "Confirm Session", Alert: alert}
		CONFIRM_SESSION_TEMPLATE.Execute(w, sessionForm)
	} else {
		posts := &DisplayPostsPage{Title: "Latest Posts", Alert: alert, Session: s, Person: p, Posts: m}
		ALL_POSTS_TEMPLATE.Execute(w, posts)
	}
}
