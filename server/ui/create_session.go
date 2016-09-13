// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"bytes"
	"database/sql"
	"github.com/Banrai/TeamWork.io/server/cryptutil"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/emailer"
	"net/http"
	"strings"
)

type CreateSessionPage struct {
	Title string
	Alert *Alert
}

func CreateSession(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	alert := new(Alert)
	alert.Message = "If you do not have a public key associated with your email address, you can <a href=\"/uploadPK\">upload it here</a>"

	if "POST" == r.Method {
		r.ParseForm()

		// an email address means create a new session, if there is at least one public key associated
		em, emExists := r.PostForm["userEmail"]
		if emExists {
			email := strings.Join(em, "")
			if len(email) > 0 {

				fn := func(stmt map[string]*sql.Stmt) {
					// attempt to find the person for this email address
					person, personErr := database.LookupPerson(stmt[database.PERSON_LOOKUP_BY_EMAIL], email)
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

					// find this person's public keys
					publicKeys, publicKeysErr := person.LookupPublicKeys(stmt[database.PK_LOOKUP])
					if publicKeysErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
					}

					if len(publicKeys) < 1 {
						alert.AlertType = "alert-warning"
						alert.Icon = "fa-hand-paper-o"
						alert.Message = "You need at least one public key associated with your email address: go <a href=\"/uploadPK\">here to upload it</a>"
						return
					}

					// go ahead and create the session
					session := new(database.SESSION)
					sessionCode, sessionCodeErr := session.Add(stmt[database.SESSION_INSERT], person.Id, SESSION_WORDS, SESSION_DURATION)
					if sessionCodeErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
					}

					// use the person's public keys to encrypt the session code
					encryptedCode, encryptedCodeErr := cryptutil.EncryptData(publicKeys, sessionCode)
					if encryptedCodeErr != nil {
						alert.AlertType = "alert-danger"
						alert.Icon = "fa-exclamation-triangle"
						alert.Message = OTHER_ERROR
						return
					}

					// send email with encryped code, and notification to html template
					uuid := cryptutil.GenerateUUID(cryptutil.UndashedUUID)
					sessionSubject := "Your TeamWork.io session"
					messageData := []string{
						"Here is your TeamWork.io session information.",
						"Decrypt the attached file with your private key, and use it at the session form."}
					attachments := []*emailer.EmailAttachment{&emailer.EmailAttachment{ContentType: emailer.TEXT_MIME, Contents: encryptedCode, FileName: uuid, FileLocation: uuid}}

					var msgBuffer bytes.Buffer
					EMAIL_TEMPLATE.Execute(&msgBuffer, &EmailMessage{Subject: sessionSubject, Heading: sessionSubject, Message: messageData})
					sendErr := emailer.Send(sessionSubject,
						msgBuffer.String(),
						emailer.TEXT_MIME,
						&emailer.EmailAddress{DisplayName: "TeamWork.io", Address: CONTACT_SENDER},
						&emailer.EmailAddress{DisplayName: person.Email, Address: person.Email},
						attachments)
					if sendErr != nil {
						http.Error(w, sendErr.Error(), http.StatusInternalServerError)
					} else {
						// present the session code form
						Redirect("/confirm")(w, r)
					}
				}

				database.WithDatabase(db, fn)
			}
		}
	}

	sessionForm := &CreateSessionPage{Title: "New Session", Alert: alert}
	CREATE_SESSION_TEMPLATE.Execute(w, sessionForm)
}
