// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package emailer

import (
	"regexp"
	"strings"
)

// Use a basic regex to confirm the given is plausibly an email address
func IsPossibleEmail(email string) bool {
	validEmail := regexp.MustCompile(`.+@.+\..+`)
	return validEmail.MatchString(strings.ToLower(email))
}
