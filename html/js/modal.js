TEAMWORK.showModal = function (title, header, message) {
    $('#modalWindow').modal('show');
    $('#modalTitle').text(title);
    $('#modalMessageHeader').text(header);
    $('#modalMessage').text(message);
}
