var $table = $('#table');

function initTable() {
    $table.bootstrapTable({
        striped: true, height: getHeight(),
        columns: [ [ 
            { field: 'crawler_name', title: 'CrawlerName', align: 'center', valign: 'middle'},
            { field: 'status', title: 'Status', align: 'center', valign: 'middle' },
            { field: 'weight', title: 'Weight', align: 'center', valign: 'middle' },
            { field: 'author', title: 'Author', align: 'center', valign: 'middle' },
            { field: 'create_time', title: 'CreateTime', align: 'center', valign: 'middle', formatter: timeFormatter },
            { field: 'modify_time', title: 'ModifyTime', align: 'center', valign: 'middle', formatter: timeFormatter },
            { field: 'operate', title: 'Operate', align: 'center', valign: 'middle', events: operateEvents, formatter: operateFormatter }
        ] ]
    });
    // sometimes footer render error.
    setTimeout(function () { $table.bootstrapTable('resetView'); }, 200);

    $table.on('expand-row.bs.table', function (e, index, row, $detail) {
        if (index % 2 == 1) {
            $detail.html('Loading from ajax request...');
            $.get('LICENSE', function (res) {
                $detail.html(res.replace(/\n/g, '<br>'));
            });
        }
    });
  
    $table.on('all.bs.table', function (e, name, args) {
        //console.log(name, args);
    });
  
    $(window).resize(function () { $table.bootstrapTable('resetView', { height: getHeight() }); });
}

function responseHandler(res) {
    $.each(res.rows, function (i, row) {
        //alert(row.author);
    });
    return res;
}

function operateFormatter(value, row, index) {
    return [
        '<a class="edit" href="javascript:void(0)" title="Edit">',
        '<i class="glyphicon glyphicon-edit"></i>',
        '</a>  ',
        '<a class="remove" href="javascript:void(0)" title="Remove">',
        '<i class="glyphicon glyphicon-remove"></i>',
        '</a>'
    ].join('');
}

function timeFormatter(value, row, index) {
    if (value <= 0) return "-";
    return new Date(value * 1000).toLocaleString();
}

$('#addnew').click(function () { window.open("/editor/"); });

window.operateEvents = {
    'click .edit': function (e, value, row, index) { window.open("/editor/?name=" + row.crawler_name); },
    'click .remove': function (e, value, row, index) {
        $.ajax({
            url: "/api/crawler/delete/" + row.crawler_name, cache: false,
            success: function(data) {
                $table.bootstrapTable('remove', { field: 'crawler_name', values: [row.crawler_name] });
            },
            error: function(XMLHttpRequest, textStatus, errorThrown) {
                $.fn.modalAlert(XMLHttpRequest.responseText, "error");
            }
        });
    }
};

function getHeight() { return $(window).height() - $('h1').outerHeight(true); }

$(function () {
    var scripts = [
            location.search.substring(1) || './bootstrap-table/bootstrap-table.js',
            './bootstrap-table/bootstrap-table-export.js',
            './bootstrap-table/tableExport.js',
            './bootstrap-table/bootstrap-table-editable.js',
            './bootstrap-table/bootstrap-editable.js'
        ],
        eachSeries = function (arr, iterator, callback) {
            callback = callback || function () {};
            if (!arr.length) {
                return callback();
            }
            var completed = 0;
            var iterate = function () {
                iterator(arr[completed], function (err) {
                    if (err) {
                        callback(err);
                        callback = function () {};
                    } else {
                        completed += 1;
                        if (completed >= arr.length) {
                            callback(null);
                        } else {
                            iterate();
                        }
                    }
                });
            };
            iterate();
        };
    eachSeries(scripts, getScript, initTable);
});

function getScript(url, callback) {
    var head = document.getElementsByTagName('head')[0];
    var script = document.createElement('script');
    script.src = url;
    var done = false;
    script.onload = script.onreadystatechange = function() {
        if (!done && (!this.readyState ||
                this.readyState == 'loaded' || this.readyState == 'complete')) {
            done = true;
            if (callback)
                callback();
            // Handle memory leak in IE
            script.onload = script.onreadystatechange = null;
        }
    };
    head.appendChild(script);
    // We handle everything using the script element injection
    return undefined;
}
