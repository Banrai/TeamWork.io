TEAMWORK.countRecipients = function () {
    return $('#recipients').find('option:selected').length;
}

TEAMWORK.selectedRecipients = function () {
    var keys = [];
    $('#recipients option:selected').each(function() {
	keys.push($(this).val());
    });
    return keys;
}

TEAMWORK.keyOwner = function (keyId) {
    var result = null;
    if( TEAMWORK.keys[keyId] ) {
	result = TEAMWORK.keys[keyId];
    }
    return result;
}

TEAMWORK.confirmRecipient = function (email) {
    var result = false;
    $('#recipients option').each(function() {
	var recipient = $(this);
	if( recipient.val() == email ) {
	    recipient.prop("selected", true);
	    $('#recipients').trigger("chosen:updated");
	    result = true;
	}
    });
    return result;
}

TEAMWORK.clearSelectedRecipients = function () {
    $('#recipients option:selected').prop("selected", false);
    $('#recipients').trigger("chosen:updated");
}

TEAMWORK.addMoreRecipients = function () {
    $('#recipient-team').show();
    $('#recipient').hide();
    $('#add-recipient').show();
    $('#message').focus();
}

TEAMWORK.reset = function () {
    $('#recipient').attr('disabled', false);
    $('#message').attr('disabled', false);
    $('#post').attr('disabled', true);
    $('#enc').attr('disabled', true);
    $('#message').val('');
    $('#recipient-search').hide();
    if( TEAMWORK.countRecipients() > 0 ) {
	TEAMWORK.addMoreRecipients();
    } else {
	$('#recipient-team').hide();
	$('#add-recipient').hide();
	$('#recipient').show();
	$('#recipient').val('');
	$('#recipient').focus();
    }	
}

TEAMWORK.showError = function (msg) {
    if( msg.endsWith("Session is expired or invalid") ) {
	TEAMWORK.showConfirmModal("Sorry", "Your session has expired", "Please click 'New Session' to get back on the saddle", "/session", "New Session");
    } else {
	TEAMWORK.showModal("Whoops!", "Sorry, but it seems we have a problem:", msg);
    }
}

$(function(){
    if( !Modernizr.formvalidation ) {
        window.location = "/browser";
    }

    $('.chosen-select').chosen({width: '100%'});
    TEAMWORK.reset();
    
    $('#recipient').keypress(function (e) {
	if( 13 === e.which ) {
	    var elem  = $(this),
		email = elem.val();
	    
	    elem.attr('disabled', true);
	    $('#recipient-search').show();

	    $.post("/searchPublicKeys",
		   { email:     email.toLowerCase(),
		     personId:  TEAMWORK.person,
		     sessionId: TEAMWORK.session})
		.done(function(reply) {
		    if( reply["msg"] && reply["err"] ) {
			TEAMWORK.showError(reply["msg"] + ": "+ reply["err"]);
		    } else if( reply["msg"] ) {
			TEAMWORK.showError(reply["msg"]);
		    } else if( reply["err"] ) {
			TEAMWORK.showError(reply["err"]);
		    } else {
			if( reply.length > 0 ) {
			    $('#recipient-team').show();
			} else {
			    // no keys found, so prompt for upload
			    TEAMWORK.showConfirmModal("Sorry", "We could not find any public keys for "+email, "But if you have a copy, you can click 'Upload Public Key' to add it yourself", "/upload", "Upload Public Key");
			}
			$.each(reply, function (i, d) { 
			    if( d["key"] ) { 
				// insert new dom PK and recipients entry
				$("body").append( $("<div class='PK' id='"+d["id"]+"' style='display:none;'>"+ d["key"] +"</div>") );
				TEAMWORK.keys[d["id"]] = email;
				if( !TEAMWORK.confirmRecipient(email) ) {
				    // some people have multiple keys per email, so don't repeat them in the selection
				    $('#recipients').append("<option value='"+email+"' selected='selected'>"+email+"</option>");
				}
				$('#recipients').trigger("chosen:updated");
			    }
			});
		    }
		    $('#recipient-search').hide();
		    elem.attr('disabled', false);
		    elem.val('');
		})
		.fail(function(reply) {
		    if( reply["msg"] && reply["err"] ) {
			TEAMWORK.showError(reply["msg"] + ": "+ reply["err"]);
		    } else if( reply["msg"] ) {
			TEAMWORK.showError(reply["msg"]);
		    } else if( reply["err"] ) {
			TEAMWORK.showError(reply["err"]);
		    } else {
			// something bad happened
			TEAMWORK.showError("");
		    }
		    $('#recipient-search').hide();
		    elem.attr('disabled', false);
		    elem.val('');
		});

	    TEAMWORK.addMoreRecipients();
	    return false;
	}
    });
    
    $('a.toggle').click(function(event){
        event.preventDefault();
        var elem     = $(this),
	    toggleId = '#'+elem.attr('href').split('#')[1];
        $(toggleId).toggle(300);
        $(toggleId).focus();
        elem.toggle();
    });

    $('#message').on('change keyup paste', function() {
	if( $('#post').is(':disabled') ) {
	    var message = $('#message').val();
	    if( message.length > 0 ) {
		$('#post').attr('disabled', false);
		$('#enc').attr('disabled', false);
	    }
	}
    });

    $('#enc').on('change', function() {
	if( $(this).is(':checked') ) {
	    $('#post').attr('disabled', true);
	    var formKeys   = [],
		message    = $('#message').val(),
		recipients = TEAMWORK.selectedRecipients();
	    if( message.length < 1 ) {
		TEAMWORK.showModal("Not so fast!", "Please type a message first", "There is nothing to encrypt");
		TEAMWORK.reset();
		TEAMWORK.addMoreRecipients();
		$(this).attr('checked', false);
	    } else if( recipients.length < 1 ) {
		TEAMWORK.showModal("Hold on!", "It's not TeamWork without others", "Please select at least one recipient first");
		TEAMWORK.reset();
		TEAMWORK.addMoreRecipients();
		$(this).attr('checked', false);
	    } else {
		$('.PK').each(function(i) {
		    var pk  = $(this).text(),
			id  = $(this).attr('id'),
			key = openpgp.key.readArmored(pk),
			usr = TEAMWORK.keyOwner(id);
		    if( $.inArray(usr, recipients) > -1 || $.inArray(id, TEAMWORK.authorKeys) > -1 ) {
			formKeys.push(key.keys[0]);
		    } 
		});

		options = {
		    data: message,
		    publicKeys: formKeys,
		    armor: true
		};

		openpgp.encrypt(options).then(function(message) {
		    $('#message').val(message.data);
		    $('#message').attr('disabled', true);
		    $('#post').attr('disabled', false);
		}, function(err) {
		    TEAMWORK.reset();
		    TEAMWORK.showError(err);
		});
	    }
	} else {
	    TEAMWORK.reset();
	}
    });

    $('#post').click(function() {
	if( TEAMWORK.countRecipients() > 0 ) {
	    $('#message').attr('disabled', false);
	    return true;
	} else {
	    TEAMWORK.showModal("Wait up!", "It seems like you forgot something:", "Please select at least one recipient first");
	    return false;
	}
    });
    
    $('#reset').click(function() {
	TEAMWORK.reset();
	TEAMWORK.clearSelectedRecipients();
    });
    
});
