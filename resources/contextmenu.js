"use strict";

function MenuDisplay(imageLoader) {
    this.menu = null;
    this.imageLoader = imageLoader;
}

MenuDisplay.prototype.display = function (x, y, menu) {
    this.dismiss();
    this.menu = new ContextMenu(x, y, menu, this.imageLoader);
};

MenuDisplay.prototype.dismiss = function () {
    if (this.menu !== null) {
        this.menu.close();
    }
};

function ContextMenu(x, y, menu, imageLoader) {
    x -= 1;
    y -= 1;
    this.imageLoader = imageLoader;
    this.submenu = null;
    this.div = document.createElement("div");
    this.div.classList.add("contextmenu");
    this.div.style.left = x + "px";
    this.div.style.top = y + "px";
    this.renderInto(menu, this.div);
    document.body.appendChild(this.div);
}

ContextMenu.prototype.renderInto = function (data, into) {
    for (let i = 0; i < data.length; i++) {
        into.appendChild(this.renderEntry(data[i]))
    }
};

ContextMenu.prototype.replaceSubMenu = function (menu) {
    if (this.submenu !== null) {
        this.submenu.close();
    }
    this.submenu = menu;
};

ContextMenu.prototype.renderEntry = function (data) {
    const entry = document.createElement("div");
    const imgBox = this.imageLoader.buildHTMLIcon(data);
    if (imgBox) {
        entry.appendChild(imgBox);
    }
    const span = document.createElement("span");
    span.textContent = data.name;
    entry.appendChild(span);
    const outerThis = this;
    if (data.contents) {
        entry.addEventListener("mouseover", function () {
            const rect = entry.getBoundingClientRect();
            const attachX = rect.x + rect.width + 1;
            const attachY = rect.y;
            outerThis.replaceSubMenu(new ContextMenu(attachX, attachY, data.contents, outerThis.imageLoader));
        });
    }
    if (data.select) {
        entry.addEventListener("click", function () {
            outerThis.close();
            data.select.call(data);
        });
    }
    return entry;
};

ContextMenu.prototype.close = function () {
    if (this.submenu !== null) {
        this.submenu.close();
        this.submenu = null;
    }
    this.div.remove();
};
