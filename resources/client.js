"use strict";

function imageLoader(images, callback) {
    var elements = {};

    function loadImage(filename, done) {
        var img = new Image();
        img.addEventListener("load", function () {
            elements[filename] = img;
            done();
        }, false);
        img.src = "img/" + filename;
    }

    var totalLoaded = 0;
    for (var i = 0; i < images.length; i++) {
        loadImage(images[i], function () {
            totalLoaded += 1;
            if (totalLoaded >= images.length) {
                callback(elements);
                callback = null;
            }
        })
    }
}

function getWebSocketURL() {
    var url = new URL("/websocket", window.location.href);
    url.protocol = (url.protocol === "http:") ? "ws:" : "wss:";
    return url.href;
}

function startWebSocket(url, message, close) {
    var reportedClose = false;
    var socket = new WebSocket(url);
    function reportClose() {
        if (!reportedClose) {
            reportedClose = true;
            close()
        }
    }
    socket.addEventListener('open', function () {
        console.log("connection opened");
    });
    socket.addEventListener('error', function () {
        console.log("connection error");
        reportClose();
    });
    socket.addEventListener('message', function (ev) {
        message(ev.data);
    });
    socket.addEventListener('close', function () {
        console.log("connection terminated");
        reportClose();
    });
}

function prepareGame(canvas) {
    var images = null;
    var isTerminated = false;
    var gameActive = false;
    var width = 640, height = 480;
    var gameSprites = {};

    function renderLoading(ctx) {
        ctx.fillStyle = 'rgb(240,240,240)';
        ctx.fillRect(0, 0, width, height);
        ctx.fillStyle = 'rgb(0,0,0)';
        ctx.textBaseline = 'middle';
        ctx.textAlign = 'center';
        var message;
        if (isTerminated) {
            if (gameActive) {
                message = "Disconnected.";
            } else {
                message = "Could not connect.";
            }
        } else if (images === null) {
            message = "Loading resources...";
        } else {
            message = "Establishing connection...";
        }
        ctx.fillText(message, width / 2, height / 2);
    }

    function renderGame(ctx) {
        ctx.fillStyle = 'rgb(0,0,0)';
        ctx.fillRect(0, 0, width, height);
        for (var spriteId in gameSprites) {
            var sprite = gameSprites[spriteId];
            if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
                var image = images[sprite.icon];
                var sw = sprite.sw || image.width;
                var sh = sprite.sh || image.height;
                ctx.drawImage(image,
                    sprite.sx || 0, sprite.sy || 0, sw, sh,
                    sprite.x, sprite.y, sprite.w || sw, sprite.h || sh);
            }
        }
    }

    function draw() {
        canvas.width = width;
        canvas.height = height;
        var ctx = canvas.getContext('2d');
        ctx.font = "24px mono";
        if (!gameActive || isTerminated) {
            renderLoading(ctx);
        } else {
            renderGame(ctx);
        }
    }

    function onMessage(data) {
        if (!gameActive) {
            gameActive = true;
        }
        var message = JSON.parse(data);
        gameSprites = message.sprites || {};
        console.log("received message", message);
    }

    function onConnectionClosed() {
        isTerminated = true;
    }

    imageLoader(["cheese.dmi", "player.dmi", "floor.dmi", "wall.dmi"], function (receivedImages) {
        images = receivedImages;
        var url = getWebSocketURL();
        console.log("connecting to", url);
        startWebSocket(url, onMessage, onConnectionClosed);
    });

    draw();
    setInterval(draw, 100);
}

window.addEventListener("load", function () {
    var canvas = document.getElementById("playspace");
    if (canvas.getContext) {
        prepareGame(canvas);
    }
});
