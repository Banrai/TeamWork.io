TEAMWORK.countRecipients = function () {
    return $('#recipients').find('option:selected').length;
}

TEAMWORK.reset = function () {
    $('#recipient').attr('disabled', false);
    $('#message').attr('disabled', false);
    $('#post').attr('disabled', true);
    $('#enc').attr('disabled', true);
    $('#message').val('');
    $('#recipient-search').hide();
    if( TEAMWORK.countRecipients() > 0 ) {
	$('#recipient-team').show();
	$('#recipient').hide();
	$('#add-recipient').show();
	$('#message').focus();
    } else {
	$('#recipient-team').hide();
	$('#add-recipient').hide();
	$('#recipient').show();
	$('#recipient').focus();
    }	
}

TEAMWORK.showError = function (msg) {
    if( msg.endsWith("Session is expired or invalid") ) {
	TEAMWORK.showConfirmModal("Sorry", "Your session has expired", "Please click 'New Session' to get back on the saddle", "/session", "New Session");
    } else {
	TEAMWORK.showModal("Sorry", "There was an error", msg);
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
			} // else: no keys found, prompt for upload
			$.each(reply, function (i, d) {
			    if( d["key"] ) {
				// insert new dom PK
				$("body").append( $("<div class='PK' style='display:none;'>"+ d["key"] +"</div>") );
				// and recipients entry
				$('#recipients').append("<option value='"+email+"' selected='selected'>"+email+"</option>");
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
	    var formKeys = [],
		message  = $('#message').val();
	    if( message.length < 1 ) {
		TEAMWORK.showError("Please type a message. There is nothing to encrypt");
		TEAMWORK.reset();
		$(this).attr('checked', false);
	    } else {
		$('.PK').each(function(i) {
		    var pk  = $(this).text(),
			key = openpgp.key.readArmored(pk);
		    formKeys.push(key.keys[0]);
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
    
    $('#reset').click(function() {
	TEAMWORK.reset();
    });
    
});
