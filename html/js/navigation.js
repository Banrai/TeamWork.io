$(function(){
    if( TEAMWORK.session !== null && TEAMWORK.person !== null ) {
	$('a.sessionLink').click(function(event){
	    event.preventDefault();
	    var action = $(this).attr('href'),
		form   = $('#navSession');
	    form.attr('action', action);
	    form.submit();
	    return false;
	});
    }
});
