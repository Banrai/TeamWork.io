TEAMWORK.showModal = function (title, header, message) {
    $('#modalWindow').modal('show');
    $('#modalTitle').text(title);
    $('#modalMessageHeader').text(header);
    $('#modalMessage').text(message);
}

TEAMWORK.showConfirmModal = function (title, header, message, continueAction) {
    var continueButton = null,
	dismissButton  = null;
    if( arguments.length > 4 ) {
	continueButton = arguments[4];
    }
    if( arguments.length > 5 ) {
	dismissButton = arguments[5];
    }
    
    $('#modalWindow').modal('show');
    $('#modalTitle').text(title);
    $('#modalMessageHeader').text(header);
    $('#modalMessage').text(message);
    $('#modalContinue').show();
    $('#modalContinue').attr('href', continueAction); 
    if( continueButton !== null ) {
	$('#modalContinueButton').text(continueButton);
    }
    if( dismissButton !== null ) {
	$('#modalDismissButton').text(dismissButton);
    }
}
