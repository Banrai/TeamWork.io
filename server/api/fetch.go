// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// For accessing public keys directly from user-provided URLs

// Retrieve the public key from the given url, and return its contents as a
// string, if it is valid
func URLFetch(url string) (string, error) {
	noKey := "" // default response

	response, respErr := http.Get(url)
	if respErr != nil {
		return noKey, respErr
	}
	if response.StatusCode != http.StatusOK {
		return noKey, errors.New(fmt.Sprintf("Error retrieving public key from '%s': %s", url, response.Status))
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return noKey, err
	}
	return string(contents), nil
}
