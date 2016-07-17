// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package cryptutil

import (
	"errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"strings"
)

// Attempt to convert the given string into an armored Block,
// returning an error if the string is not valid armor
func DecodeArmoredKey(key string) (*armor.Block, error) {
	block, err := armor.Decode(strings.NewReader(key))
	if nil == block {
		return block, errors.New("Invalid armored text")
	}
	return block, err
}

// Convert the armored key into an OpenPGP Entity
func AsEntity(key string) (*openpgp.Entity, error) {
	decoded, decodedErr := DecodeArmoredKey(key)
	if decodedErr != nil {
		return nil, decodedErr
	}

	packets := packet.NewReader(decoded.Body)
	return openpgp.ReadEntity(packets)
}
