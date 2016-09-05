// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
)

type CreateSessionPage struct {
	Title   string
  Alert *Alert
}

func CreateSession(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	if "POST" == r.Method {
		r.ParseForm()
		// check for session/person validity
	}

	sessionForm := &CreateSessionPage{Title: "New Session", Alert: &Alert{Message: "If you do not have a public key associated with your email address, you can <a href=\"/uploadPK\">upload it here</a>"}}
	CREATE_SESSION_TEMPLATE.Execute(w, sessionForm)
}
