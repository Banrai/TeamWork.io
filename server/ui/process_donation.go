// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"net/http"
	"strconv"
)

type DonationPage struct {
	Title    string
	Alert    *Alert
	Session  *database.SESSION
	Person   *database.PERSON
	StripePK string
}

func ProcessDonation(w http.ResponseWriter, r *http.Request, db database.DBConnection, opts ...interface{}) {
	var (
		s  *database.SESSION
		p  *database.PERSON
		pk string
		sk string
	)
	alert := new(Alert)

	// fetch the stripe public (pk) and secret (sk) keys
	for i, k := range opts {
		switch i {
		case 0:
			pk = fmt.Sprintf("%v", k)
		case 1:
			sk = fmt.Sprintf("%v", k)
		}
	}

	if "POST" == r.Method {
		r.ParseForm()
		attemptCharge := true

		if len(sk) == 0 {
			alert.AsError("Bad configuration: missing the stripe secret key")
			attemptCharge = false
		}

		stripe.Key = sk

		token := r.PostForm.Get("stripeToken")
		if len(token) == 0 {
			alert.AsError("Bad request: no stripe checkout token")
			attemptCharge = false
		}

		amount := r.PostForm.Get("amount")
		if len(amount) == 0 {
			alert.AsError("Amount missing")
			attemptCharge = false
		}

		changeAmount, changeAmountErr := strconv.ParseUint(amount, 10, 64)
		if changeAmountErr != nil {
			alert.AsError("Amount invalid")
			attemptCharge = false
		}

		if attemptCharge {
			chargeParams := &stripe.ChargeParams{
				Amount:   (changeAmount * 100),
				Currency: "usd",
				Desc:     "Donation to TeamWork.io",
			}
			chargeParams.SetSource(token)
			_, err := charge.New(chargeParams)
			if err == nil {
				// success
				alert.AlertType = "alert-success"
				alert.Icon = "fa-heart"
				alert.Message = "Thank You for Your Donation!"
			} else {
				alert.AsError(err.Error())
			}
		}

	}

	if s == nil && p == nil {
		s = new(database.SESSION)
		p = new(database.PERSON)
	}

	donationForm := &DonationPage{Title: TITLE_DONATE, Alert: alert, Session: s, Person: p, StripePK: pk}
	DONATE_TEMPLATE.Execute(w, donationForm)
}
