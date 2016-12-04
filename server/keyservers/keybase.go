// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package keyservers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// For accessing public keys from KeyBase (https://keybase.io/)

// Lookup the public key for the given username, and return it as a string,
// if it is valid
func KeyBaseSearch(userName string) (string, error) {
	noKey := "" // default response

	response, respErr := http.Get(fmt.Sprintf("https://keybase.io/%s/key.asc", userName))
	if respErr != nil {
		return noKey, respErr
	}
	if response.StatusCode != http.StatusOK {
		return noKey, errors.New(fmt.Sprintf("Error retrieving public key for KeyBase user '%s': %s", userName, response.Status))
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return noKey, err
	}
	return string(contents), nil
}
