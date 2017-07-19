
function GetQueryString(name) {
  var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
  var r = window.location.search.substr(1).match(reg);
  if (r != null) return unescape(r[2]);
  return null;
}

// Initialize the editor with a JSON schema
var editor = new JSONEditor(document.getElementById('editor_holder'), {
  ajax: true, disable_properties: true,
  schema: { $ref: "./schema/crawler_item.json" }, startval: null
});

var indicator = document.getElementById('valid_indicator');

function submit() {
    var errors = editor.validate();
    if (errors.length) {
        indicator.style.color = 'red';
        var err = errors[0].path + ": " + errors[0].message;
        indicator.textContent = err;
        return
    }
    var name = GetQueryString("name");
    var url = "/api/crawler/create/";
    if (name != null) {
      url = "/api/crawler/update/";
    }
    $.ajax({
      type: "POST",
      url: url + editor.getValue().crawler_name,
      dataType: "json",
      data: JSON.stringify(editor.getValue()),
      success: function(data) {
        indicator.style.color = 'green';
        indicator.textContent = JSON.stringify(data);
      },
      error: function(XMLHttpRequest, textStatus, errorThrown) {
        indicator.style.color = 'red';
        indicator.textContent = XMLHttpRequest.responseText;
      }
    });
}

document.getElementById('submit').addEventListener('click', submit);

editor.on('change',function() {
  var errors = editor.validate();
  if (errors.length) {
    indicator.style.color = 'red';
    indicator.textContent = "not valid";
  } else {
    indicator.style.color = 'green';
    indicator.textContent = "valid";
  }
});

function fillEditor(url) {
  $.getJSON(url, function(result) {
    editor.setValue(result);
    //console.log(JSON.stringify(result));
  });
}

document.getElementById('restore').addEventListener('click',function() {
  fillEditor("default.json");
});

editor.on('ready', function(){
  var name = GetQueryString("name");
  if (name != null) {
    console.log(name);
    fillEditor("/api/crawler/retrieve/" + name);
  }
});
