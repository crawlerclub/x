$(document).ready(function() {
    var name = GetQueryString("name");
    if (name != null) {
        $('#text').val(name);
        testCrawler(name);
    }
    $('#submit').click(function() {
        var name = $('#text').val().trim();
        if (name != "") {
            testCrawler(name);
        }
    });
});

function testCrawler(name) {
    $('#editor_holder').html("<h4>loading...</h4>");
    $.ajax({
        url: "/api/test/"+name, cache: false,
        success: function(result) {
            $('#editor_holder').jsonview(result);
        },
        error: function(XMLHttpRequest, textStatus, errorThrown) {
            alert(XMLHttpRequest.responseText);
        }
    });
}
