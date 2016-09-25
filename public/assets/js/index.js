/* 
 * Copyright (C) 2016 andylizi
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
function asyncLoadScript(url, callback) {
    function preload(url) {
        var img = new Image();
        img.src = url;
        url = img.src;
        delete img;
        return url;
    }
    if (!(url instanceof Array)) {
        var $script = $("<script></script>").attr({
            type: "text/javascript",
            defer: "defer",
            async: "async"
        }), script = $script[0];
        if (callback && typeof callback == "function" && script.readyState) {
            $script.one("readystatechange", function () {
                if (script.readyState == "loaded" || script.readyState == "complete") {
                    return callback();
                }
            });
        } else {
            $script.one("load", callback);
        }
        return $("head")[0].appendChild($script.attr("src", url)[0]);
    }
    var scripts = $("script");
    for (var i = 0; i < url.length; i++) {
        var fullURL = preload(url[i]);
        scripts.each(function () {
            if ((this.src == url[i]) || (this.src == fullURL)) {
                url.splice(i, 1, "");
            }
        });
    }
    if (url.length == 0) {
        if (callback && typeof callback == "function") {
            callback();
        }
        return false;
    }
    var finishCount = 0;
    for (var i = 0; i < url.length; i++)
        if (url[i])
            asyncLoadScript(url[i], function () {
                if (++finishCount == url.length)
                    callback();
            });
}
function login() {
    var finished = false;
    if ($("#username").html() == "Anonymous") {
        setTimeout(function () {
            if (!finished) {
                $("#l_loading").fadeIn("slow");
            }
        }, 200);
        if ($("#login_form")) {
            asyncLoadScript(["assets/js/lib/message.js", "assets/js/lib/jquery.md5.js"], function () {
                $("#login").load("ajax/login.jsp", show);
            });
        } else
            show();
        function show() {
            finished = true;
            $("#l_loading").fadeOut("slow");
            $("#btn_login").html("[Cancel]").one("click", function () {
                $("#login").slideUp("slow");
                $("#btn_login").html("[Sign in]").one("click", login);
                $("#actions").slideDown("slow");
            });
            $("#login").slideDown("slow");
        }
    } else {
        setTimeout(function () {
            if (!finished) {
                $("#username_arena").text("Loading...");
            }
        }, 200);
        $("#username_arena").load("ajax/login.jsp", {action: "loginout"}, function () {
            finished = true;
            $("#login,#l_loading").fadeOut("slow");
        });
    }
}
function vibrate(e, callback) {
    var ir = e.css("margin-right");
    var il = e.css("margin-left");
    var ol = e.css("border-bottom");
    e.css("border-bottom", "1px red solid").animate({
        marginRight: "+10px",
        marginLeft: "-10px"
    }, 50).animate({
        marginRight: "-10px",
        marginLeft: "+10px"
    }, 50).animate({
        marginRight: "+10px",
        marginLeft: "-10px"
    }, 50).animate({
        marginRight: "-10px",
        marginLeft: "+10px"
    }, 50, function () {
        e[0].focus();
        e.css({marginRight: ir, marginLeft: il});
        e.one("keyup", function () {
            e.css("border-bottom", ol)[0].focus();
        });
        if (callback) {
            callback(e);
        }
    });
}
(function () {
    function showJoin() {
        if ($("#btn_back:visible").length > 0)
            return;
        location.hash = "#join";
        $("#btn_create").animate({
            "margin-right": "10px"
        }, "slow");
        $("#txt_join").css("visibility", "visible");
        $("#btn_back").fadeIn("slow").one("click", function () {
            location.hash = "";
            $(this).fadeOut("normal", function () {
                setTimeout(function () {
                    $("#btn_join").prop("disabled", false);
                    setTimeout(function () {
                        $("#btn_create").css("visibility", "visible").fadeIn("slow");
                    }, 200);
                    $("#txt_join").animate({
                        width: "0px"
                    }, "slow").fadeOut("fast", function () {
                        $(this).css("display", "inline").css("visibility", "hidden")[0].focus();
                        $("#txt_join").val("");
                    });
                    $("#btn_join").one("click", showJoin).animate({
                        marginLeft: "20px"
                    }, "fast");
                }, 100);
            });
        });
        $("#btn_join").prop("disabled", !$("#txt_join").val());
        setTimeout(function () {
            $("#txt_join").animate({
                width: "150px"
            }, "slow", "swing", function () {
                this.focus();
            }).keyup(function (e) {
                $("#btn_join").prop("disabled", !this.value);
                if (!this.value) {
                    return;
                }
                if (e.keyCode == 13) {
                    $("#btn_join").click();
                }
            });
        }, 200);
        $("#btn_join").css("margin-left", "10px").on("click", function () {
            var val = $("#txt_join").val();
            if (!val) {
                $("#txt_join")[0].focus();
                return;
            }
            $("#txt_join").val((val = val.replace(/^#/, "")));
            $("#btn_join").val("Loading...");
            $.get("ajax/join.jsp?" + val, function (data) {
                if (parseInt(data) == 1) {
                    location.href = "game?" + val;
                } else {
                    vibrate($("#txt_join"));
                }
            });
        });
    }
    var socket = new Socket();
    function initStatusSocket(connectCount) {
        if (!connectCount) {
            connectCount = 1;
        }
        if (connectCount > 3)
            return;
        var list = $("#list ul");
        var loc = window.location;
        var socketuri = "ws://";
        if(loc.protocol == "https:") {
            socketuri = "wss://";
        }
        socketuri += loc.host;
        var path = loc.pathname.substr(0,(loc.pathname.lastIndexOf("/")));
        socketuri += path;
        socketuri += "/status";
        socket.connect(socketuri, function () {}, function (evt) {
            list.empty();
            if (evt.data == "{}") {
                list.parent(":visible").slideUp("slow");
                return;
            }
            $.each($.parseJSON(evt.data), function (id, v) {
                list.append("<li>Room <span class='id'>#" + id + "</span>&nbsp;" + v.owner + "&nbsp;&nbsp;<span class='" + (v.playing ? "playing" : "waiting") + "'>[" +
                        (v.playing ? "Playing" : "Waiting") + "]</span>&nbsp;\n\
                                    Rounds: " + v.rounds + "&nbsp;&nbsp;\n\
                                    Steps: " + v.steps + "&nbsp;&nbsp;\n\
                                    Watchers: " + v.watchers + "&nbsp;&nbsp;&nbsp;&nbsp;<a href='game?" + id + "'>[" + (v.playing ? "Watch" : "Join") + "]</a></li>");
            });
            $(".id").click(function () {
                $("#txt_join").val(this.innerHTML.replace(/^#/, ""));
                showJoin();
            });
            list.parent(":hidden").slideDown("slow");
        }, function () {}, function () {
            initStatusSocket(++connectCount);
        });
    }

    $(function () {
        $("#title").css("background-image", "url('assets/image/bgs/" + parseInt(Math.random() * 6) + ".svg')");
        if ($("#main").width() < 890) {
            $("#gear").hide();
        }
        $("#main").css({
            display: "none",
            visibility: "visible"
        }).fadeIn("slow");
        $("#btn_create").click(function () {
            $("#btn_join").fadeOut("fast");
            $(this).children("span").text("Loading...");
            location.href = "game?create";
        });
        $("#btn_list").click(function () {
            $("#btn_list .icon").toggleClass("turn_off").toggleClass("turn_on");
        });
        $("#btn_join").one("click", showJoin);
        $("#username_arena").load("ajax/login.jsp", {action: "info"});
        if (location.hash == "#join") {
            showJoin();
        }
        initStatusSocket();
    });
})(window.jQuery);
