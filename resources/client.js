"use strict";

function prepareGame(canvas, inputsource, verbentry, paneltabs, panelbody, textoutput) {
    let gameActive = false;
    const imageLoader = new ImageLoader("resource");
    const statPanels = new StatPanel(paneltabs, panelbody, imageLoader);
    const contextMenu = new MenuDisplay(imageLoader);
    const session = new Session();
    const render = new Canvas(canvas, imageLoader);
    const soundPlayer = new SoundPlayer();
    const keyHandler = new KeyHandler(inputsource);

    function sendVerb(verb) {
        console.log("send verb", verb);
        session.sendMessage({"verb": verb})
    }

    keyHandler.onmove = function (direction) {
        sendVerb("." + direction);
    };

    function getLoadingMessage() {
        if (session.terminated) {
            if (gameActive) {
                return "Disconnected.";
            } else {
                return "Could not connect.";
            }
        } else if (!imageLoader.isLoaded()) {
            return "Loading resources...";
        } else {
            return "Establishing connection...";
        }
    }

    function draw() {
        if (!gameActive || session.terminated) {
            render.renderLoading(getLoadingMessage());
        } else {
            keyHandler.tick();
            render.renderGame();
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

    function spriteContextMenu(sprite) {
        const menu = [];
        if (sprite.verbs && sprite.verbs.length > 0) {
            for (let j = 0; j < sprite.verbs.length; j++) {
                menu.push({
                    "name": sprite.verbs[j],
                    "targetUID": sprite.uid,
                    "select": function () {
                        // FIXME: when "targetUID" gets large enough, it will stop being able to be accurately represented by javascript's numbers
                        sendVerb(this.name + " #" + this.targetUID);
                    },
                });
            }
        }
        return menu;
    }

    statPanels.oncontextmenu = function (ev, entry) {
        const menu = spriteContextMenu(entry);
        if (menu.length !== 0) {
            contextMenu.display(ev.pageX, ev.pageY, menu);
        }
    };

    statPanels.onverb = function (verb) {
        sendVerb(verb);
    };

    session.onmessage = function (message) {
        if (!gameActive) {
            gameActive = true;
        }
        if (message.newstate) {
            render.updateSprites(message.newstate.sprites || []);
            render.updateSize(message.newstate.viewportwidth, message.newstate.viewportheight);
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
                soundPlayer.playSound(sound);
            }
        }
    };

    session.onclose = function () {
        soundPlayer.cancelAllSounds();
    };

    inputsource.addEventListener("contextmenu", function (ev) {
        ev.preventDefault();
    });

    inputsource.addEventListener("click", function () {
        contextMenu.dismiss();
    });

    canvas.addEventListener("contextmenu", function (ev) {
        contextMenu.dismiss();
        const sprites = render.findSprites(ev);
        const menu = [];
        for (let i = 0; i < sprites.length; i++) {
            const sprite = sprites[i];
            const contents = spriteContextMenu(sprite);
            if (contents.length > 0) {
                menu.push({
                    "name": sprite.name,
                    "icon": sprite.icon,
                    "frames": sprite.frames,
                    "sw": sprite.sw,
                    "sh": sprite.sh,
                    "contents": contents,
                });
            }
        }
        if (menu.length > 0) {
            contextMenu.display(ev.pageX, ev.pageY, menu);
        }
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

    if (resources === undefined) {
        console.log("expected resources.js to be included for resource list");
        return;
    }
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
