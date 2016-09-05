// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"github.com/Banrai/TeamWork.io/server/database"
	"net/http"
)

type NewPostPage struct {
	Title   string
	Session *database.SESSION
	Person  *database.PERSON
	Keys    []*database.PUBLIC_KEY
}

func PostMessage(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	if "POST" == r.Method {
		r.ParseForm()
		// check for session/person validity
	}

	postForm := &NewPostPage{Title: "New Post", Session: &database.SESSION{}, Person: &database.PERSON{}}
	NEW_POST_TEMPLATE.Execute(w, postForm)
}
