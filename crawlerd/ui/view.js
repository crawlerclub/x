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
    $.ajax({
        url: "/api/crawler/retrieve/"+name, cache: false,
        success: function(result) {
            $('#editor_holder').jsonview(result);
        },
        error: function(XMLHttpRequest, textStatus, errorThrown) {
            alert(XMLHttpRequest.responseText);
        }
    });
}
