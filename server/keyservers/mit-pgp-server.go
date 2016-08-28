// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package keyservers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

/* For accessing public keys from the
   MIT PGP Public Key Server
	 http://pgp.mit.edu/
*/

const MIT_SOURCE = "http://pgp.mit.edu/"

// parse the specific key links from a request of the form:
// http://pgp.mit.edu/pks/lookup?search=me@example.org excluding revoked/not
// verified keys
func parseKeyResult(in io.Reader) (string, error) {
	dec := xml.NewDecoder(in)
	captureData := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if "pre" == tok.Name.Local {
				captureData = true
			}
		case xml.EndElement:
			captureData = false
		case xml.CharData:
			if captureData {
				return strings.TrimSpace(fmt.Sprintf("%s", tok)), nil
			}
		}
	}
	return "", nil
}

// parse the specific key links from a request of the form:
// http://pgp.mit.edu/pks/lookup?search=me@example.org
// excluding revoked/not verified keys
func parseMatchResult(in io.Reader) ([]string, error) {
	dec := xml.NewDecoder(in)
	var results []string
	captureData := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return results, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if "a" == tok.Name.Local {
				for _, attr := range tok.Attr {
					if "href" == attr.Name.Local {
						if strings.HasPrefix(attr.Value, "/pks/lookup?op=get&search=") {
							captureData = true
							results = append(results, html.UnescapeString(attr.Value))
						}
					}
				}
			}
		case xml.CharData:
			if captureData {
				text := strings.TrimSpace(fmt.Sprintf("%s", tok))
				if strings.IndexAny(text, "not verified") > -1 ||
					strings.IndexAny(text, "REVOKED") > -1 {
					// pop the most recent link, since it is invalid
					results = results[:len(results)-1]
				}
				captureData = false
			}
		}
	}
	return results, nil
}

// fetch the contents of the given url using http get, and return the
// contents as an io.Reader object
func getLinkContents(url string) (io.Reader, error) {
	response, respErr := http.Get(url)
	if respErr != nil {
		return nil, respErr
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return bytes.NewReader(contents), err
}

// search the MIT PGP Server for all public keys corresponding to this
// email address, returning them as armored strings, excluding revoked/not
// verified keys
func MITSearch(email string) ([]string, error) {
	var keys []string

	in, inErr := getLinkContents(fmt.Sprintf("http://pgp.mit.edu/pks/lookup?search=%s", email))
	if inErr != nil {
		return keys, inErr
	}

	links, linkErr := parseMatchResult(in)
	if linkErr != nil {
		return keys, linkErr
	}

	for _, link := range links {
		pkIn, pkInErr := getLinkContents(fmt.Sprintf("http://pgp.mit.edu%s", link))
		if pkInErr != nil {
			return keys, pkInErr
		}

		key, keyErr := parseKeyResult(pkIn)
		if keyErr != nil {
			return keys, keyErr
		}

		keys = append(keys, key)
	}

	return keys, nil
}
