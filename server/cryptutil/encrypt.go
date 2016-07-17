// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package cryptutil

import (
	"bytes"
	"github.com/Banrai/TeamWork.io/server/database"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

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
