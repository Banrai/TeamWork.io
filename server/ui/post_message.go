// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/database"
	"log"
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
	var (
		s *database.SESSION
		p *database.PERSON
		k []*database.PUBLIC_KEY
	)
	alert := new(Alert)
	alert.Message = "You need to create a session to be able to post a new message. If you have already decryted a session code, you can <a href=\"/confirm\">confirm it here</a>"
	confirmSession := false

	if "POST" == r.Method {
		r.ParseForm()
		confirmSession = true

		sessionCode, sessionCodeExists := r.PostForm["session"]
		personCode, personCodeExists := r.PostForm["person"]
		if sessionCodeExists && personCodeExists {
			sessionId := strings.Join(sessionCode, "")
			personId := strings.Join(personCode, "")
			if len(sessionId) > 0 && len(personId) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					// make sure the session is valid
					session, sessionErr := ConfirmSessionCode(sessionId, stmt[database.SESSION_CLEANUP], stmt[database.SESSION_LOOKUP_BY_ID])
					if sessionErr != nil {
						alert.AsError(OTHER_ERROR)
						return
					}

					if len(session.Id) == 0 {
						alert.AsError(INVALID_SESSION)
						return
					}

					// attempt to find the person for this session
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_ID], session.PersonId)
					log.Println("")
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

					// make sure the session matches the person from the form
					if person.Id != personId {
						alert.AsError(INVALID_SESSION)
						return
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

					// session, person, and keys are valid
					s = session
					p = person
					k = keys

					// see if there is a message to post
					messageData, messageDataExists := r.PostForm["message"]
					if !messageDataExists {
						alert.Message = "Please write a message before posting (encryption is optional)"
						return
					}

					// post the message to the database
					message := new(database.MESSAGE)
					message.Message = strings.Join(messageData, "")
					message.PersonId = person.Id
					msgId, msgIdErr := message.Add(stmt[database.MESSAGE_INSERT], MESSAGE_DURATION)
					if msgIdErr != nil {
						alert.AsError("Your message could not be posted at this time")
						return
					}
					message.Id = msgId

					// add the list of recipients to the message
					recipientList, recipientListExists := r.PostForm["recipients"]
					if recipientListExists {
						people := make([]*database.PERSON, 0)
						for _, recipientEmail := range recipientList {
							recipient, recipientErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_EMAIL], recipientEmail)
							if recipientErr != nil { // for now, but maybe prepare an alert
								log.Println(recipientErr)
							} else {
								people = append(people, recipient)
							}
						}
						if len(people) > 0 {
							for _, msgRecipientErr := range message.AddRecipients(stmt[database.RECIPIENT_INSERT], people) {
								if msgRecipientErr != nil { // for now
									log.Println(msgRecipientErr)
								}
							}
						}
					}

					// success
					alert.Message = "Your message has been posted"
				}
				database.WithDatabase(db, fn)
			}
		}
	}

	if confirmSession && s == nil && p == nil {
		sessionForm := &ConfirmSessionPage{Title: "Confirm Session", Alert: alert}
		CONFIRM_SESSION_TEMPLATE.Execute(w, sessionForm)
	} else {
		postPage := &NewPostPage{Title: "New Post", Alert: alert, Session: s, Person: p, Keys: k}
		NEW_POST_TEMPLATE.Execute(w, postPage)
	}

}
