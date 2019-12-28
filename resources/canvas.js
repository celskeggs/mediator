"use strict";

function Canvas(canvas, imageLoader) {
    this.canvas = canvas;
    this.imageLoader = imageLoader;
    this.width = 672;
    this.height = 672;
    this.aspectRatio = this.width / this.height;
    this.aspectShiftX = this.aspectShiftY = 0;
}

Canvas.prototype.startRender = function (fill) {
    this.canvas.width = this.width;
    this.canvas.height = this.height;
    var ctx = this.canvas.getContext('2d');
    ctx.font = "24px mono";
    ctx.fillStyle = fill;
    ctx.fillRect(0, 0, this.width, this.height);
    return ctx;
};

Canvas.prototype.renderLoading = function (message) {
    const ctx = this.startRender('rgb(240,240,240)');
    ctx.fillStyle = 'rgb(0,0,0)';
    ctx.textBaseline = 'middle';
    ctx.textAlign = 'center';
    ctx.fillText(message, this.width / 2, this.height / 2);
};

Canvas.prototype.renderGame = function (sprites) {
    const ctx = this.startRender('rgb(0,0,0)');
    for (let i = 0; i < sprites.length; i++) {
        const sprite = sprites[i];
        if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
            const image = this.imageLoader.getImage(sprite.icon);
            if (!image) {
                continue;
            }
            const sw = sprite.sw || image.width;
            const sh = sprite.sh || image.height;
            const drawW = sprite.w || sw;
            const drawH = sprite.h || sh;
            const drawX = this.aspectShiftX + sprite.x;
            const drawY = this.aspectShiftY + this.height - sprite.y - drawH;
            ctx.drawImage(image,
                sprite.sx || 0, sprite.sy || 0, sw, sh,
                drawX, drawY, drawW, drawH);
        }
    }
};

Canvas.prototype.findSprites = function (ev, gameSprites) {
    const pos = this.getMousePosition(ev);
    const sprites = [];
    for (let i = 0; i < gameSprites.length; i++) {
        const sprite = gameSprites[i];
        if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
            const image = this.imageLoader.getImage(sprite.icon);
            if (!image) {
                continue;
            }
            const sw = sprite.sw || image.width;
            const sh = sprite.sh || image.height;
            const drawW = sprite.w || sw;
            const drawH = sprite.h || sh;
            const drawX = this.aspectShiftX + sprite.x;
            const drawY = this.aspectShiftY + this.height - sprite.y - drawH;
            if (pos.x >= drawX && pos.y >= drawY && pos.x < drawX + drawW && pos.y < drawY + drawH) {
                sprites.push(sprite);
            }
        }
    }
    return sprites;
};

Canvas.prototype.updateSize = function (newwidth, newheight) {
    if (!newwidth || !newheight) {
        return;
    }
    if (newheight * this.aspectRatio > newwidth) {
        this.width = Math.round(newheight * this.aspectRatio);
        this.height = newheight;
        this.aspectShiftX = Math.floor((width - newwidth) / 2);
        this.aspectShiftY = 0;
    } else if (newwidth / this.aspectRatio > newheight) {
        this.width = newwidth;
        this.height = Math.round(newwidth / this.aspectRatio);
        this.aspectShiftX = 0;
        this.aspectShiftY = Math.floor((height - newheight) / 2);
    } else {
        this.width = newwidth;
        this.height = newheight;
        this.aspectShiftX = this.aspectShiftY = 0;
    }
};

Canvas.prototype.getMousePosition = function (ev) {
    const rect = this.canvas.getBoundingClientRect();
    let x = ev.clientX - rect.left;
    let y = ev.clientY - rect.top;
    x = x / rect.width * this.width;
    y = y / rect.height * this.height;
    return {x: x, y: y};
};
