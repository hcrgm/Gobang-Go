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
$(function () {
    $(document.body).on("mousedown","button",function (event) {
        var $this = $(this);
        if(!$this.children().hasClass("wrapper")){
            $this.wrapInner("<span></span>");
        }
        var offset = $this.offset();
        var $ripple = $('<div></div>').addClass('ripple').prependTo($this).css({
            left: event.pageX - offset.left,
            top: event.pageY - offset.top
        });
        setTimeout(function () {
            $ripple.css("transform","scale(280)");
        }, 1);
    }).on("mouseup","button",function () {
        var $this = $(this).children('.ripple').css('opacity', 0);
        setTimeout(function () {
            $this.remove();
        }, 1000);
    });
});
