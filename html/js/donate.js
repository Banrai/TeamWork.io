TEAMWORK.resetDonation = function () {
    $('#amount').val('');
    $("#amount").focus();
}

$(function(){
    TEAMWORK.resetDonation();

    var handler = StripeCheckout.configure({
	key: TEAMWORK.stripePK,
	locale: 'auto',
	name: 'Donate to TeamWork.io',
	description: 'Charges will appear as "Banrai LLC"',
	token: function(token) {
	    $('input#stripeToken').val(token.id);
	    $('form').submit();
	}
    });

    $("#amountCheck").on('click', function(event) {
	event.preventDefault();

	var amount = $('#amount').val();
	amount = amount.replace(/\$/g, '').replace(/\,/g, '')
	amount = parseFloat(amount);
	
	if (isNaN(amount)) {
	    TEAMWORK.showModal("Sorry, bad amount", "Please enter a valid amount", "You will be charged in USD");
	    TEAMWORK.resetDonation();
	} else if (amount < 10.00) {
	    TEAMWORK.showModal("Sorry, insufficient amount", "The minimum donation is $10", "You will be charged in USD");
	    TEAMWORK.resetDonation();
	} else {
	    amount = amount * 100;
	    handler.open({
		amount: Math.round(amount)
	    })
	}
    });      
});
