TEAMWORK.focusUpload = function () {
    $("#publicKeyUrl").val();
    $("#key-url").hide(300);
    $("#key-upload").show(300);
}

TEAMWORK.focusUrl = function () {
    $("#selectedFile").val();
    $("#key-upload").hide(300);
    $("#key-url").show(300);
    $("#publicKeyUrl").focus();
}

$(function(){
    if( TEAMWORK.session === null && TEAMWORK.person === null ) {
	$("#userEmail").focus();
    }
    $("#publicKey").on('change', function() {
        var input = $(this),
            label = input.val().replace(/\\/g, '/').replace(/.*\//, '');
	$("#selectedFile").html('<i class="fa fa-file-text-o"></i> '+label);
	$("#selectedFile").css('padding-top', '0.5em');
    });
    $("#keyTypeUpload").on('change', function() {
	if( $(this).is(':checked') ) {
	    TEAMWORK.focusUpload();
	} else {
	    TEAMWORK.focusUrl();
	}
    });
    $("#keyTypeURL").on('change', function() {
	if( $(this).is(':checked') ) {
	    TEAMWORK.focusUrl();
	} else {
	    TEAMWORK.focusUpload();
	}
    });
});
