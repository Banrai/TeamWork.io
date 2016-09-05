// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"time"
)

const (
	// configuration
	SESSION_WORDS    = 6 // it's a magic number
	SESSION_DURATION = 30 * time.Minute

	// Errors and alerts
	DISABLED    = "Sorry, this email address and all of its public keys has been disabled"
	OTHER_ERROR = "Sorry, there was a problem"

	// sending automated emails
	CONTACT_SENDER = "noreply@teamwork.io"
)

var (
	TEMPLATE_LIST = func(templatesFolder string, templateFiles []string) []string {
		t := make([]string, 0)
		for _, f := range templateFiles {
			t = append(t, path.Join(templatesFolder, f))
		}
		return t
	}

	UNSUPPORTED_TEMPLATE_FILE = "browser_not_supported.html"

	NEW_POST_TEMPLATE_FILES = []string{"new-post.html", "head.html", "modal.html", "scripts.html", "session.html"}
	NEW_POST_TEMPLATE       *template.Template

	CREATE_SESSION_TEMPLATE_FILES = []string{"create-session.html", "head.html", "alert.html", "scripts.html"}
	CREATE_SESSION_TEMPLATE       *template.Template

	EMAIL_TEMPLATE_FILES = []string{"email.html"}
	EMAIL_TEMPLATE       *template.Template

	TEMPLATES_INITIALIZED = false
)

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

type Alert struct {
	AlertType string
	Icon      string
	Message   template.HTML
}

type EmailMessage struct {
	Subject string
	Heading string
	Message []string
}

// InitializeTemplates confirms the given folder string leads to the html
// template files, otherwise templates.Must() will complain
func InitializeTemplates(folder string) {
	NEW_POST_TEMPLATE = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, NEW_POST_TEMPLATE_FILES)...))
	CREATE_SESSION_TEMPLATE = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, CREATE_SESSION_TEMPLATE_FILES)...))
	EMAIL_TEMPLATE = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, EMAIL_TEMPLATE_FILES)...))
	TEMPLATES_INITIALIZED = true
}
