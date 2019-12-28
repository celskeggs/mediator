"use strict";

function Canvas(canvas, imageLoader) {
    this.canvas = canvas;
    this.imageLoader = imageLoader;
    // placeholder values
    this.viewWidth = this.viewHeight = 100;
    this.width = this.height = 100;
    this.aspectShiftX = this.aspectShiftY = 0;
    this.scaleFactor = 1;
    this.frameID = 1;
    this.gameSprites = [];
    this.animationInfo = {};
}

Canvas.prototype.updateSprites = function (sprites) {
    this.gameSprites = sprites;
};

Canvas.prototype.startRender = function (fill) {
    const rect = this.canvas.getBoundingClientRect();
    this.canvas.width = rect.width;
    this.canvas.height = rect.height;
    this.updateSizing();
    const ctx = this.canvas.getContext('2d');
    ctx.font = "24px mono";
    ctx.fillStyle = fill;
    ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
    return ctx;
};

Canvas.prototype.renderLoading = function (message) {
    const ctx = this.startRender('rgb(240,240,240)');
    ctx.fillStyle = 'rgb(0,0,0)';
    ctx.textBaseline = 'middle';
    ctx.textAlign = 'center';
    ctx.fillText(message, this.canvas.width / 2, this.canvas.height / 2);
};

Canvas.prototype.prepareRenderImage = function (sprite, animationInfo, frameID) {
    const info = this.imageLoader.prepareImage(sprite, animationInfo, frameID);
    if (!info) {
        return null;
    }
    // low and high corners computed in floating point
    const lx = this.aspectShiftX + sprite.x * this.scaleFactor;
    const ly = this.canvas.height - this.aspectShiftY - (sprite.y + info.dh) * this.scaleFactor;
    const hx = lx + info.dw * this.scaleFactor;
    const hy = ly + info.dh * this.scaleFactor;
    // round positions before converting them back to position/size
    info.dx = Math.round(lx);
    info.dy = Math.round(ly);
    info.dw = Math.round(hx) - Math.round(lx);
    info.dh = Math.round(hy) - Math.round(ly);
    return info;
};

Canvas.prototype.renderGame = function () {
    const ctx = this.startRender('rgb(0,0,0)');
    ctx.imageSmoothingEnabled = false;
    for (let i = 0; i < this.gameSprites.length; i++) {
        const sprite = this.gameSprites[i];
        if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
            const info = this.prepareRenderImage(sprite, this.animationInfo, this.frameID);
            if (!info) {
                continue;
            }
            ctx.drawImage(info.img,
                info.sx, info.sy, info.sw, info.sh,
                info.dx, info.dy, info.dw, info.dh);
        }
    }
    this.frameID += 1;
};

Canvas.prototype.findSprites = function (ev) {
    const pos = this.getMousePosition(ev);
    const sprites = [];
    for (let i = 0; i < this.gameSprites.length; i++) {
        const sprite = this.gameSprites[i];
        if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
            const info = this.prepareRenderImage(sprite, null, null);
            if (info && pos.x >= info.dx && pos.y >= info.dy && pos.x < info.dx + info.dw && pos.y < info.dy + info.dh) {
                sprites.push(sprite);
            }
        }
    }
    return sprites;
};

Canvas.prototype.updateSizing = function () {
    const aspectRatio = this.canvas.width / this.canvas.height;
    if (this.viewHeight * aspectRatio > this.viewWidth) {
        this.width = Math.round(this.viewHeight * aspectRatio);
        this.height = this.viewHeight;
        this.scaleFactor = (this.canvas.height / this.viewHeight);
        this.aspectShiftX = Math.floor((this.width - this.viewWidth) * this.scaleFactor / 2);
        this.aspectShiftY = 0;
    } else if (this.viewWidth / aspectRatio > this.viewHeight) {
        this.width = this.viewWidth;
        this.height = Math.round(this.viewWidth / aspectRatio);
        this.scaleFactor = (this.canvas.width / this.viewWidth);
        this.aspectShiftX = 0;
        this.aspectShiftY = Math.floor((this.height - this.viewHeight) * this.scaleFactor / 2);
    } else {
        this.width = this.viewWidth;
        this.height = this.viewHeight;
        this.scaleFactor = (this.canvas.width / this.viewWidth);
        this.aspectShiftX = this.aspectShiftY = 0;
    }
};

Canvas.prototype.updateSize = function (newwidth, newheight) {
    if (!newwidth || !newheight) {
        return;
    }
    this.viewWidth = newwidth;
    this.viewHeight = newheight;
    this.updateSizing();
};

Canvas.prototype.getMousePosition = function (ev) {
    const rect = this.canvas.getBoundingClientRect();
    let x = ev.clientX - rect.left;
    let y = ev.clientY - rect.top;
    return {x: x, y: y};
};
