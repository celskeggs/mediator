"use strict";

function KeyHandler(inputSource) {
    this.keyDirection = null;
    this.onmove = null;

    const handler = this;

    inputSource.addEventListener("keydown", function (ev) {
        handler.press(ev);
    });

    inputSource.addEventListener("keyup", function (ev) {
        handler.release(ev);
    });
}

KeyHandler.prototype.press = function (ev) {
    const direction = this.arrowKeyDirection(ev);
    if (direction) {
        ev.preventDefault();
        this.keyDirection = direction;
    }
};

KeyHandler.prototype.release = function (ev) {
    const direction = this.arrowKeyDirection(ev);
    if (direction) {
        ev.preventDefault();
        if (this.keyDirection === direction) {
            this.keyDirection = null;
        }
    }
};

KeyHandler.prototype.arrowKeyDirection = function (ev) {
    if (ev.code === "ArrowUp") {
        return "north";
    } else if (ev.code === "ArrowDown") {
        return "south";
    } else if (ev.code === "ArrowLeft") {
        return "west";
    } else if (ev.code === "ArrowRight") {
        return "east";
    } else {
        return null;
    }
};

KeyHandler.prototype.tick = function () {
    if (this.keyDirection !== null && this.onmove !== null) {
        this.onmove(this.keyDirection);
    }
};
