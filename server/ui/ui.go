// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"io/ioutil"
	"net/http"
	"path"
)

var UNSUPPORTED_TEMPLATE_FILE = "browser_not_supported.html"

// Use this to redirect one request to another target (string)
func Redirect(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// Respond to requests using HTML templates and the standard Content-Type (i.e., "text/html")
func MakeHTMLHandler(fn func(http.ResponseWriter, *http.Request, database.DBConnection, ...interface{}), db database.DBConnection, opts ...interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db, opts...)
	}
}

// Show the static template for unsupported browsers
func UnsupportedBrowserHandler(templatesFolder string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadFile(path.Join(templatesFolder, UNSUPPORTED_TEMPLATE_FILE))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(body))
	}
}

// handlers for static resources
func StaticFolder(folder string, templatesFolder string) http.Handler {
	return http.StripPrefix(fmt.Sprintf("/%s/", folder), http.FileServer(http.Dir(path.Join(templatesFolder, fmt.Sprintf("../%s/", folder)))))
}
