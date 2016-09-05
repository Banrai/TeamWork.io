// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package cryptutil

import (
	"bytes"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"os"
)

// Universally Unique Identifier (UUID) creation

var (
	// Formatting function: take the 16 random bytes and return them as a string
	// of 8-4-4-4-12 tuples, separated by dashes
	DashedUUID = func(b []byte) string {
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	}

	// An alternative formatting function: take the 16 random bytes and return
	// them as a single string, with no dashes
	UndashedUUID = func(b []byte) string {
		return fmt.Sprintf("%x", b)
	}
)

// Generate a universally unique identifier (UUID) using the computer's
// /dev/urandom output as a randomizer, returning a string specified by
// the given formatting function
func GenerateUUID(fn func([]byte) string) string {
	f, e := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	defer f.Close()

	if e != nil {
		return ""
	} else {
		b := make([]byte, 16)
		f.Read(b)
		return fn(b)
	}
}

// Encrypt the data with the list of PGP keys, returning an armored string
func EncryptData(keys []*database.PUBLIC_KEY, data string) (string, error) {
	pgpKeys := []*openpgp.Entity{}
	for _, key := range keys {
		pgpKey, pgpKeyErr := AsEntity(key.Key)
		if pgpKeyErr != nil {
			return "", pgpKeyErr
		}
		pgpKeys = append(pgpKeys, pgpKey)
	}

	encbuf := bytes.NewBuffer(nil)
	w, wErr := armor.Encode(encbuf, "PGP MESSAGE", nil)
	if wErr != nil {
		return "", wErr
	}

	plaintext, plaintextErr := openpgp.Encrypt(w, pgpKeys, nil, nil, nil)
	if plaintextErr != nil {
		return "", plaintextErr
	}

	_, writeErr := plaintext.Write([]byte(data))
	if writeErr != nil {
		return "", writeErr
	}
	plaintext.Close()
	w.Close()

	return encbuf.String(), nil
}
