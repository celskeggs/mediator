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
    if (image === "") {
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
