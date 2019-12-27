"use strict";

function MenuDisplay() {
    this.menu = null;
}

MenuDisplay.prototype.display = function (x, y, menu, icons) {
    this.dismiss();
    this.menu = new ContextMenu(x, y, menu, icons);
};

MenuDisplay.prototype.dismiss = function () {
    if (this.menu !== null) {
        this.menu.close();
    }
};

function ContextMenu(x, y, menu, icons) {
    x -= 1;
    y -= 1;
    this.icons = icons;
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
    if (data.icon) {
        const imgBox = document.createElement("div");
        imgBox.style.overflow = "hidden";
        const baseImg = this.icons[data.icon];
        imgBox.style.width = (data.sw || baseImg.width) + "px";
        imgBox.style.height = (data.sh || baseImg.height) + "px";
        const img = baseImg.cloneNode(true);
        img.marginLeft = "-" + (data.sx || 0) + "px";
        img.marginTop = "-" + (data.sy || 0) + "px";
        imgBox.appendChild(img);
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
            outerThis.replaceSubMenu(new ContextMenu(attachX, attachY, data.contents, outerThis.items));
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
