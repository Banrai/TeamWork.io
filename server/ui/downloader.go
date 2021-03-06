// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"io"
	"net/http"
	"strings"
)

const NO_SUCH_MESSAGE = "There is no such message among the list of available posts"

// Lookup and stream the given message back to the client
func DownloadMessage(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	var (
		m      *database.MESSAGE
		s      *database.SESSION
		p      *database.PERSON
		d      []*database.MESSAGE_DIGEST
		domain string
	)
	alert := new(Alert)
	messageFound := false

	// get the server domain
	for i, k := range opts {
		switch i {
		case 0:
			domain = fmt.Sprintf("%v", k)
		}
	}

	messageId := r.URL.Query().Get("message")

	// preserve the session data, if any
	if "POST" == r.Method {
		r.ParseForm()
		sessionCode, sessionCodeExists := r.PostForm["session"]
		personCode, personCodeExists := r.PostForm["person"]
		if sessionCodeExists && personCodeExists {
			sessionId := strings.Join(sessionCode, "")
			personId := strings.Join(personCode, "")
			if len(sessionId) > 0 && len(personId) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					session, sessionErr := ConfirmSessionCode(sessionId, stmt[database.SESSION_CLEANUP], stmt[database.SESSION_LOOKUP_BY_ID])
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

					// make sure the session matches the person from the form
					if person.Id != personId {
						alert.AsError(INVALID_SESSION)
						return
					}

					// found a valid session and person
					s = session
					p = person
				}
				database.WithDatabase(db, fn)
			}
		}
	}

	if len(messageId) > 0 {
		// attempt to find the specific message
		fn := func(stmt map[string]*sql.Stmt) {
			messages, messagesErr := database.RetrieveMessages(stmt[database.MESSAGE_BY_ID], messageId, 1, 0)
			if messagesErr != nil {
				alert.AsError(OTHER_ERROR)
				return
			}
			if len(messages) > 0 {
				m = messages[0]
				messageFound = true
			}
		}
		database.WithDatabase(db, fn)
	}

	if messageFound {
		// send its contents to the client
		var contents string
		if len(domain) > 0 {
			contents = strings.Replace(m.Message, "http://openpgpjs.org", fmt.Sprintf("%s using openpgpjs.org", domain), 1)
		} else {
			contents = m.Message
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.asc", messageId))
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(contents)))
		io.Copy(w, bytes.NewBufferString(contents))
	} else {
		// go back to the list of posts
		alert.AsError(NO_SUCH_MESSAGE)

		if s == nil && p == nil {
			// retrieve the latest digests, without session/person
			fn := func(stmt map[string]*sql.Stmt) {
				messages, _ := database.RetrieveMessages(stmt[database.LATEST_MESSAGES], "", POSTS_PER_PAGE, 0)
				digests, _ := database.GetMessageDigests(stmt[database.PERSON_LOOKUP_BY_ID], stmt[database.RECIPIENTS_BY_MESSAGE], messages, "")
				d = digests
			}
			database.WithDatabase(db, fn)

			// define these as empty, so the session template renders properly
			s = new(database.SESSION)
			p = new(database.PERSON)
		} else {
			// use the session data
			fn := func(stmt map[string]*sql.Stmt) {
				messages, _ := p.LookupLatestMessages(stmt[database.LATEST_MESSAGES], POSTS_PER_PAGE, 0)
				digests, _ := database.GetMessageDigests(stmt[database.PERSON_LOOKUP_BY_ID], stmt[database.RECIPIENTS_BY_MESSAGE], messages, p.Id)
				d = digests
			}
			database.WithDatabase(db, fn)
		}

		posts := &DisplayPostsPage{Title: "Latest Posts", Alert: alert, Session: s, Person: p, Posts: d}
		ALL_POSTS_TEMPLATE.Execute(w, posts)
	}
}
