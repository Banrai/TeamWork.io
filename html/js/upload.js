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
});
