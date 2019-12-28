"use strict";

/* ImageLoader loads <img> elements for each of a number of icons.
 * The public interface:
 *  - load(iconlist)
 *  - isLoaded()
 *  - getImage(name)
 * The events to be overridden by the user:
 *  - onload
 */

function ImageLoader(basepath) {
    this.started = false;
    this.images = {};
    this.pending = 0;
    basepath = basepath || "";
    if (basepath && !basepath.endsWith("/")) {
        basepath += "/";
    }
    this.basepath = basepath;
    this.onload = null;
}

ImageLoader.prototype.loadImage = function (filename) {
    const loader = this;
    const img = new Image();
    img.addEventListener("load", function () {
        loader.receive(filename, img);
    }, false);
    img.src = this.basepath + filename;
};

ImageLoader.prototype.receive = function (filename, img) {
    if (this.pending <= 0) {
        throw new Error("pending image load count underflow");
    }
    this.images[filename] = img;
    this.pending -= 1;
    if (this.pending === 0 && this.onload !== null) {
        this.onload();
    }
};

ImageLoader.prototype.isLoaded = function () {
    return this.started && this.pending === 0;
};

ImageLoader.prototype.load = function (images) {
    if (this.started) {
        // this might be a reasonable operation; we just don't support it yet
        throw new Error("attempt to reuse image loader");
    }
    this.started = true;
    this.pending = images.length;
    for (let i = 0; i < images.length; i++) {
        this.loadImage(images[i]);
    }
};

ImageLoader.prototype.getImage = function (image) {
    if (!image) {
        return null;
    }
    if (!this.isLoaded()) {
        console.log("attempt to use icon before icons were loaded");
        return null;
    }
    const loaded = this.images[image];
    if (!loaded) {
        console.log("attempt to use icon that was never loaded:", image);
        return null;
    }
    return loaded;
};

ImageLoader.prototype.prepareImage = function (sprite) {
    const image = this.getImage(sprite.icon);
    if (!image) {
        return null;
    }
    const sx = sprite.sx || 0;
    const sy = sprite.sy || 0;
    const sw = sprite.sw || image.width;
    const sh = sprite.sh || image.height;
    const drawW = sprite.w || sw;
    const drawH = sprite.h || sh;
    return {
        "img": image,
        "sx": sx,
        "sy": sy,
        "sw": sw,
        "sh": sh,
        "dw": drawW,
        "dh": drawH,
    };
};

ImageLoader.prototype.updateHTMLIcon = function (sprite, imgBox) {
    const info = this.prepareImage(sprite);
    if (!info) {
        if (imgBox.children.length > 0) {
            imgBox.children[0].remove();
        }
        imgBox.style.width = "";
        imgBox.style.height = "";
        imgBox.style.overflow = "";
        return null;
    } else {
        imgBox.style.width = info.sw + "px";
        imgBox.style.height = info.sh + "px";
        imgBox.style.overflow = "hidden";
        if (imgBox.children.length > 0 && imgBox.children[0].src !== info.img.src) {
            imgBox.children[0].remove();
        }
        if (imgBox.children.length === 0) {
            imgBox.appendChild(info.img.cloneNode(true));
        }
        imgBox.children[0].marginLeft = "-" + info.sx + "px";
        imgBox.children[0].marginTop = "-" + info.sy + "px";
        return imgBox;
    }
};

ImageLoader.prototype.buildHTMLIcon = function (sprite) {
    const imgBox = document.createElement("div");
    return this.updateHTMLIcon(sprite, imgBox);
};
