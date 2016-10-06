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
window.Message = new (window.MessageAPI = Class.extend({
    init: function (close_callback, back_callback) {
        $(function () {
            if ($("#mask").length <= 0) {
                $(document.body).append('<div id="mask" style="display:none;opacity:0;"></div>');
            }
            if ($("#mask #dialog").length <= 0) {
                $("#mask").append('<div id="dialog" style="margin-top:50px;opacity:0;"><div id="dialog_title"></div><div id="dialog_content"></div><div id="dialog_actions"></div></div>');
            }
        });
        this.dialogQueue = [];
        this.snackbarQueue = [];
        this.dialogOpened = false;
        this.close_callback = close_callback;
        this.back_callback = back_callback;
    },
    _openDialog_: function (title, content, actions, callback, opt) {
        var self = this;
        if (opt) {
            setTimeout(function () {
                self._open_(title, content, actions, callback);
            }, 500);
            return;
        }
        console.log("Open dialog - " + title);
        this.dialogOpened = true;
        $("#dialog_title").html(title);
        $("#dialog_content").html(content);
        if (actions) {
            $("#dialog_actions").html(actions);
            if (this.close_callback && typeof this.close_callback == "function") {
                $("#btn_close").click(this.close_callback);
            }
            if (this.back_callback && typeof this.close_callback == "function") {
                $("#btn_back").click(this.back_callback);
            }
        } else {
            $("#dialog_actions").empty();
        }
        $("#mask").css("display", "block");
        setTimeout(function () {
            $("#mask").css("opacity", "1");
            $("#dialog").css({
                opacity: 1
            });
            setTimeout(function () {
                $("#dialog").css({
                    marginTop: "90px"
                });
                if (callback)
                    if (typeof callback == "function")
                        return callback();
                    else if (typeof callback == "string")
                        return eval(callback);
            }, 5);
        }, 5);
    },
    openDialog: function (title, content, actions, callback, top) {
        if (this.dialogOpened) {
            if (!top) {
                console.log("Buffer: dialog - " + title);
                this.dialogQueue.push({title: title, content: content, actions: actions, callback: callback});
                return;
            } else {
                this.__param_title_ = title;
                this.__param_content_ = content;
                this.__param_actions = actions;
                this.__param_callback = callback;
                return this.closeDialog("this._openDialog_(this.__param_title_,this.__param_content_,this.__param_actions,this.__param_callback,true);", false);
            }
        }
        return this._openDialog_(title, content, actions, callback);
    },
    closeDialog: function (callback, next) {
        var self = this;
        console.log("Close dialog");
        self.dialogOpened = false;
        $("#mask").css("opacity", 0);
        setTimeout(function () {
            $("#mask").css("display", "none");
        }, 450);
        $("#dialog").css({
            marginTop: "50px",
            opacity: 0
        });
        if (next || (typeof next) == "undefined") {
            setTimeout(function () {
                if (self.dialogQueue.length != 0) {
                    var task = self.dialogQueue.shift();
                    self.openDialog(task.title, task.content, task.actions, task.callback);
                    return;
                }
            }, 500);
        }
        if (callback)
            if (typeof callback == "function")
                return callback();
            else if (typeof callback == "string")
                return eval(callback);
    },
    makeSnackbar: function (content, time, callback) {
        if (!content)
            return;
        if ($('.snackbar').length > 0) {
            this.snackbarQueue.push({content: content, time: time, callback: callback});
            return false;
        }
        var $snackbar = $('<div class="snackbar"></div>').html(content).wrapInner('<div></div>').appendTo(document.body);
        var self = this;
        setTimeout(function () {
            $snackbar.css({
                visibility: 'visible',
                transform: 'translate3d(0px, 0px, 0px)'
            });
            setTimeout(function () {
                $snackbar.css({
                    visibility: 'hidden',
                    transform: 'translate3d(0px, ' + $snackbar.height() + 'px, 0px)'
                });
                setTimeout(function () {
                    $snackbar.remove();
                    if (callback)
                        if (typeof callback == 'function')
                            callback();
                        else if (typeof callback == 'string')
                            eval(callback);
                    if (self.snackbarQueue.length != 0) {
                        var task = self.snackbarQueue.shift();
                        self.makeSnackbar(task.content, task.time, task.callback);
                    }
                }, 600);
            }, time ? time : 5000);
        }, 10);
        return true;
    },
    isOpen: function () {
        return this.dialogOpened;
    },
    onCloseButton: function (close_callback) {
        this.close_callback = typeof close_callback == "function" ? close_callback : null;
    },
    onBackButton: function (back_callback) {
        this.back_callback = typeof back_callback == "function" ? back_callback : null;
    }
}));