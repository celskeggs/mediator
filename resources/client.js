"use strict";

function prepareGame(canvas, inputsource, verbentry, paneltabs, panelbody, textoutput) {
    var gameActive = false;
    var width = 672, height = 672;
    var aspectRatio = width / height;
    var aspectShiftX = 0, aspectShiftY = 0;
    var gameSprites = [];
    var keyDirection = null;
    const imageLoader = new ImageLoader("resource");
    const statPanels = new StatPanel(paneltabs, panelbody, imageLoader);
    const contextMenu = new MenuDisplay(imageLoader);
    const session = new Session();
    var Player = new SoundPlayer();

    function renderLoading(ctx) {
        ctx.fillStyle = 'rgb(240,240,240)';
        ctx.fillRect(0, 0, width, height);
        ctx.fillStyle = 'rgb(0,0,0)';
        ctx.textBaseline = 'middle';
        ctx.textAlign = 'center';
        var message;
        if (session.terminated) {
            if (gameActive) {
                message = "Disconnected.";
            } else {
                message = "Could not connect.";
            }
        } else if (!imageLoader.isLoaded()) {
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
                const image = imageLoader.getImage(sprite.icon);
                if (!image) {
                    continue;
                }
                var sw = sprite.sw || image.width;
                var sh = sprite.sh || image.height;
                var drawW = sprite.w || sw;
                var drawH = sprite.h || sh;
                var drawX = aspectShiftX + sprite.x;
                var drawY = aspectShiftY + height - sprite.y - drawH;
                ctx.drawImage(image,
                    sprite.sx || 0, sprite.sy || 0, sw, sh,
                    drawX, drawY, drawW, drawH);
            }
        }
    }

    function sendVerb(verb) {
        console.log("send verb", verb);
        session.sendMessage({"verb": verb})
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
        if (!gameActive || session.terminated) {
            renderLoading(ctx);
        } else {
            handleKeys();
            renderGame(ctx);
        }
    }

    function updateWidthHeight(newwidth, newheight) {
        if (!newwidth || !newheight) {
            return;
        }
        if (newheight * aspectRatio > newwidth) {
            width = Math.round(newheight * aspectRatio);
            height = newheight;
            aspectShiftX = Math.floor((width - newwidth) / 2);
            aspectShiftY = 0;
        } else if (newwidth / aspectRatio > newheight) {
            width = newwidth;
            height = Math.round(newwidth / aspectRatio);
            aspectShiftX = 0;
            aspectShiftY = Math.floor((height - newheight) / 2);
        } else {
            width = newwidth;
            height = newheight;
            aspectShiftX = aspectShiftY = 0;
        }
    }

    function displayText(line) {
        var shouldScroll = textoutput.scrollHeight - textoutput.scrollTop === textoutput.clientHeight;
        var nextLine = document.createElement("p");
        nextLine.textContent = line;
        textoutput.append(nextLine);
        if (shouldScroll) {
            textoutput.scrollTop = textoutput.scrollHeight - textoutput.clientHeight;
        }
    }

    // FIXME: can this be merged with the other context-menu-opening code on the regular canvas?
    statPanels.oncontextmenu = function(ev, entry) {
        if (entry.verbs && entry.verbs.length > 0) {
            const menu = [];
            for (let j = 0; j < entry.verbs.length; j++) {
                menu.push({
                    "name": entry.verbs[j],
                    "targetName": entry.name,
                    "select": function () {
                        // FIXME: uniquely identify atoms, rather than using names
                        sendVerb(this.name + " " + this.targetName);
                    },
                });
            }
            contextMenu.display(ev.pageX, ev.pageY, menu);
        }
    };

    statPanels.onverb = function (verb) {
        sendVerb(verb);
    };

    session.onmessage = function(message) {
        if (!gameActive) {
            gameActive = true;
        }
        if (message.newstate) {
            gameSprites = message.newstate.sprites || [];
            updateWidthHeight(message.newstate.viewportwidth, message.newstate.viewportheight);
            if (message.newstate.windowtitle) {
                document.getElementsByTagName("title")[0].textContent = message.newstate.windowtitle;
            }
            statPanels.update(message.newstate.verbs || [], message.newstate.stats.panels || {});
        }
        if (message.textlines) {
            for (var i = 0; i < message.textlines.length; i++) {
                displayText(message.textlines[i]);
            }
        }
        if (message.sounds) {
            for (var j = 0; j < message.sounds.length; j++) {
                var sound = message.sounds[j];
                Player.playSound(sound);
            }
        }
    };

    session.onclose = function() {
        Player.cancelAllSounds();
    };

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
            ev.preventDefault();
        }
    });

    inputsource.addEventListener("keyup", function (ev) {
        var direction = keyCodeToDirection(ev.code);
        if (direction === keyDirection) {
            keyDirection = null;
            ev.preventDefault();
        }
    });

    inputsource.addEventListener("contextmenu", function (ev) {
        ev.preventDefault();
    });

    inputsource.addEventListener("click", function () {
        contextMenu.dismiss();
    });

    function canvasMousePosition(ev) {
        const rect = canvas.getBoundingClientRect();
        var x = ev.clientX - rect.left;
        var y = ev.clientY - rect.top;
        x = x / rect.width * width;
        y = y / rect.height * height;
        return {x: x, y: y};
    }

    function findSprites(x, y) {
        var sprites = [];
        for (var i = 0; i < gameSprites.length; i++) {
            var sprite = gameSprites[i];
            if (sprite.icon && sprite.x !== undefined && sprite.y !== undefined) {
                const image = imageLoader.getImage(sprite.icon);
                if (!image) {
                    continue;
                }
                var sw = sprite.sw || image.width;
                var sh = sprite.sh || image.height;
                var drawW = sprite.w || sw;
                var drawH = sprite.h || sh;
                var drawX = aspectShiftX + sprite.x;
                var drawY = aspectShiftY + height - sprite.y - drawH;
                if (x >= drawX && y >= drawY && x < drawX + drawW && y < drawY + drawH) {
                    sprites.push(sprite);
                }
            }
        }
        return sprites;
    }

    canvas.addEventListener("contextmenu", function (ev) {
        contextMenu.dismiss();
        var pos = canvasMousePosition(ev);
        var sprites = findSprites(pos.x, pos.y);
        var menu = [];
        for (var i = 0; i < sprites.length; i++) {
            var sprite = sprites[i];
            if ((sprite.verbs || []).length === 0) {
                continue;
            }
            var contents = [];
            for (var j = 0; j < sprite.verbs.length; j++) {
                contents.push({
                    "name": sprite.verbs[j],
                    "targetName": sprite.name,
                    "select": function () {
                        // TODO: uniquely identify atoms, rather than using names
                        sendVerb(this.name + " " + this.targetName);
                    },
                });
            }
            menu.push({
                "name": sprite.name,
                "icon": sprite.icon,
                "sx": sprite.sx,
                "sy": sprite.sy,
                "sw": sprite.sw,
                "sh": sprite.sh,
                "contents": contents,
            });
        }
        if (menu.length > 0) {
            contextMenu.display(ev.pageX, ev.pageY, menu);
        }
        ev.preventDefault();
    });

    verbentry.focus();

    verbentry.addEventListener("keypress", function (ev) {
        if (ev.key === "Enter" && verbentry.value !== "") {
            sendVerb(verbentry.value);
            verbentry.value = "";
        }
    });

    imageLoader.onload = function () {
        session.connect();
    };

    imageLoader.load(resources);

    draw();
    setInterval(draw, 100);
}

window.addEventListener("load", function () {
    var canvas = document.getElementById("playspace");
    var verbEntry = document.getElementById("verb");
    var panelTabs = document.getElementById("paneltabs");
    var panelBody = document.getElementById("panelbody");
    var textOutput = document.getElementById("textspace");
    if (canvas.getContext) {
        prepareGame(canvas, document.body, verbEntry, panelTabs, panelBody, textOutput);
    }
});
