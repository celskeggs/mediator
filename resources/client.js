"use strict";

function imageLoader(images, callback) {
    var elements = {};

    function loadImage(filename, done) {
        var img = new Image();
        img.addEventListener("load", function () {
            elements[filename] = img;
            done();
        }, false);
        img.src = "resource/" + filename;
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

function startWebSocket(url, open, message, close) {
    var reportedClose = false;
    var socket = new WebSocket(url);

    function reportClose() {
        if (!reportedClose) {
            reportedClose = true;
            close()
        }
    }

    function sendMessage(message) {
        socket.send(JSON.stringify(message));
    }

    socket.addEventListener('open', function () {
        console.log("connection opened");
        open(sendMessage);
    });
    socket.addEventListener('error', function () {
        console.log("connection error");
        reportClose();
    });
    socket.addEventListener('message', function (ev) {
        message(JSON.parse(ev.data));
    });
    socket.addEventListener('close', function () {
        console.log("connection terminated");
        reportClose();
    });
}

function prepareGame(canvas, inputsource) {
    var images = null;
    var isTerminated = false;
    var gameActive = false;
    var width = 640, height = 480;
    var gameSprites = [];
    var keyDirection = null;
    var sendMessage = function (message) {
    };

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
        for (var i = 0; i < gameSprites.length; i++) {
            var sprite = gameSprites[i];
            if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
                var image = images[sprite.icon];
                if (!image) {
                    console.log("no such icon:", sprite.icon);
                    continue;
                }
                var sw = sprite.sw || image.width;
                var sh = sprite.sh || image.height;
                var drawW = sprite.w || sw;
                var drawH = sprite.h || sh;
                var drawX = sprite.x;
                var drawY = height - sprite.y - drawH;
                ctx.drawImage(image,
                    sprite.sx || 0, sprite.sy || 0, sw, sh,
                    drawX, drawY, drawW, drawH);
            }
        }
    }

    function sendVerb(verb) {
        console.log("send verb", verb);
        sendMessage({"verb": verb})
    }

    function handleKeys() {
        if (keyDirection != null) {
            sendVerb("." + keyDirection)
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
            handleKeys();
            renderGame(ctx);
        }
    }

    function onConnectionOpen(send) {
        sendMessage = send;
    }

    function onMessage(message) {
        if (!gameActive) {
            gameActive = true;
        }
        gameSprites = message.sprites || [];
    }

    function onConnectionClosed() {
        isTerminated = true;
    }

    if (resources === undefined) {
        console.log("expected resources.js to be included for resource list");
        return;
    }

    function keyCodeToDirection(code) {
        if (code === "ArrowUp") {
            return "north";
        } else if (code === "ArrowDown") {
            return "south";
        } else if (code === "ArrowLeft") {
            return "west";
        } else if (code === "ArrowRight") {
            return "east";
        } else {
            return null;
        }
    }

    inputsource.addEventListener("keydown", function (ev) {
        var direction = keyCodeToDirection(ev.code);
        if (direction !== null) {
            keyDirection = direction;
        }
    });

    inputsource.addEventListener("keyup", function (ev) {
        var direction = keyCodeToDirection(ev.code);
        if (direction === keyDirection) {
            keyDirection = null;
        }
    });

    imageLoader(resources, function (receivedImages) {
        images = receivedImages;
        var url = getWebSocketURL();
        console.log("connecting to", url);
        startWebSocket(url, onConnectionOpen, onMessage, onConnectionClosed);
    });

    draw();
    setInterval(draw, 100);
}

window.addEventListener("load", function () {
    var canvas = document.getElementById("playspace");
    if (canvas.getContext) {
        prepareGame(canvas, document.body);
    }
});
