// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package httputil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// For accessing URLs via HTTP GET
func getUrl(url string) ([]byte, error) {
	noData := []byte{} // default, in case of error

	response, respErr := http.Get(url)
	if respErr != nil {
		return noData, respErr
	}
	if response.StatusCode != http.StatusOK {
		return noData, errors.New(fmt.Sprintf("Error retrieving '%s' via HTTP GET: %s", url, response.Status))
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return noData, err
	}

	return contents, nil
}

// Retrieve the public key from the given url, and return its contents as a
// string, if it is valid
func URLFetchAsString(url string) (string, error) {
	noKey := "" // default response

	b, err := getUrl(url)
	if err != nil {
		return noKey, err
	}

	return string(b), nil
}

// fetch the contents of the given url using http get, and return the
// contents as an io.Reader object
func URLFetchAsReader(url string) (io.Reader, error) {
	b, err := getUrl(url)
	return bytes.NewReader(b), err
}
