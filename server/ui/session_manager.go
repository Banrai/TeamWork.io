// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/cryptutil"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/emailer"
	"time"
)

// Create a new session for this Person and email them the corresponding session code to decrypt
func CreateNewSession(person *database.PERSON, keys []*database.PUBLIC_KEY, sessionInsert *sql.Stmt) error {
	// generate a random session code for this person
	session := new(database.SESSION)
	sessionCode, sessionCodeErr := session.Add(sessionInsert, person.Id, SESSION_WORDS, SESSION_DURATION)
	if sessionCodeErr != nil {
		return sessionCodeErr
	}

	// use the person's public keys to encrypt the session code
	encryptedCode, encryptedCodeErr := cryptutil.EncryptData(keys, sessionCode)
	if encryptedCodeErr != nil {
		return encryptedCodeErr
	}

	// send email with encryped code, and notification to html template
	sessionFilename := fmt.Sprintf("TeamWork.io-session-%s.asc", time.Now().UTC().Format(time.RFC3339))
	sessionSubject := "Your TeamWork.io session"
	messageData := []string{
		"Here is your TeamWork.io session information.",
		"Decrypt the attached file with your private key, and use it at the session form."}
	attachments := []*emailer.EmailAttachment{&emailer.EmailAttachment{ContentType: emailer.TEXT_MIME, Contents: encryptedCode, FileName: sessionFilename, FileLocation: sessionFilename}}

	var textBody, htmlBody bytes.Buffer
	EMAIL_TEMPLATE.Execute(&textBody, &EmailMessage{Subject: sessionSubject, Message: messageData})
	HTML_EMAIL_TEMPLATE.Execute(&htmlBody, &EmailMessage{Subject: sessionSubject, Heading: sessionSubject, Message: messageData})
	return emailer.Send(sessionSubject,
		textBody.String(),
		htmlBody.String(),
		&emailer.EmailAddress{DisplayName: "TeamWork.io", Address: CONTACT_SENDER},
		&emailer.EmailAddress{DisplayName: person.Email, Address: person.Email},
		attachments)
}

// Wipe any expired sessions, and then confirm this code, returning the session object
func ConfirmSessionCode(code string, cleanup, lookup *sql.Stmt) (*database.SESSION, error) {
	database.CleanupSessions(cleanup)
	return database.LookupSession(lookup, code)
}

// Generate a new public key, and associate it with this person
func AddPublicKey(person *database.PERSON, keyData, keySource, keyNickname string, pkInsert *sql.Stmt) error {
	publicKey := new(database.PUBLIC_KEY)
	publicKey.Key = keyData
	publicKey.Source = keySource
	publicKey.Nickname = keyNickname
	_, pkErr := publicKey.Add(pkInsert, person.Id)
	return pkErr
}
