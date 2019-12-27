"use strict";

function prepareGame(canvas, inputsource, verbentry, paneltabs, panelbody, textoutput) {
    var gameActive = false;
    var width = 672, height = 672;
    var aspectRatio = width / height;
    var aspectShiftX = 0, aspectShiftY = 0;
    var gameSprites = [];
    var statPanels = {};
    var verbs = [];
    var selectedStatPanel = null;
    var keyDirection = null;
    const imageLoader = new ImageLoader("resource");
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

    function renderPanelTabs(tabs, selected) {
        while (paneltabs.children.length > tabs.length) {
            paneltabs.children[paneltabs.children.length-1].remove();
        }
        while (paneltabs.children.length < tabs.length) {
            var button = document.createElement("button");
            button.addEventListener("click", function () {
                selectedStatPanel = this.textContent;
                updateStatPanels();
            });
            paneltabs.appendChild(button);
        }
        for (var i = 0; i < paneltabs.children.length; i++) {
            paneltabs.children[i].textContent = tabs[i];
            var isSelected = tabs[i] === selected;
            if (isSelected) {
                paneltabs.children[i].classList.add("selected");
            } else {
                paneltabs.children[i].classList.remove("selected");
            }
        }
    }

    function renderPanelData(entries, areVerbs) {
        if (areVerbs) {
            if (!panelbody.classList.contains("verbpanel")) {
                while (panelbody.children.length > 0) {
                    panelbody.children[0].remove();
                }
            }
            panelbody.classList.add("verbpanel");
            panelbody.classList.remove("statpanel");
        } else {
            if (!panelbody.classList.contains("statpanel")) {
                while (panelbody.children.length > 0) {
                    panelbody.children[0].remove();
                }
            }
            panelbody.classList.add("statpanel");
            panelbody.classList.remove("verbpanel");
        }
        while (panelbody.children.length > entries.length) {
            panelbody.children[panelbody.children.length-1].remove();
        }
        while (panelbody.children.length < entries.length) {
            var child = document.createElement("div");
            if (!areVerbs) {
                var statlabel = document.createElement("span");
                statlabel.classList.add("statlabel");
                child.appendChild(statlabel);
                var object = document.createElement("div");
                object.appendChild(document.createElement("div"));
                object.appendChild(document.createElement("span"));
                object.entry = null;
                object.addEventListener("contextmenu", function (ev) {
                    if (this.entry !== null && this.entry.verbs && this.entry.verbs.length > 0) {
                        var menu = [];
                        for (var j = 0; j < this.entry.verbs.length; j++) {
                            menu.push({
                                "name": this.entry.verbs[j],
                                "targetName": this.entry.name,
                                "select": function () {
                                    // TODO: uniquely identify atoms, rather than using names
                                    sendVerb(this.name + " " + this.targetName);
                                },
                            });
                        }
                        contextMenu.display(ev.pageX, ev.pageY, menu);
                    }
                });
                child.appendChild(object);
                var statsuffix = document.createElement("span");
                statsuffix.classList.add("statsuffix");
                child.appendChild(statsuffix);
            } else {
                child.addEventListener("click", function (e) {
                    sendVerb(this.textContent);
                });
            }
            panelbody.appendChild(child);
        }
        for (var i = 0; i < panelbody.children.length; i++) {
            var data = entries[i];
            var entry = panelbody.children[i];
            if (areVerbs) {
                entry.textContent = data;
            } else {
                var labelSpan = entry.children[0];
                var statObject = entry.children[1];
                var iconDiv = statObject.children[0];
                var nameSpan = statObject.children[1];
                var suffixSpan = entry.children[2];

                if (data.verbs && data.verbs.length > 0) {
                    statObject.classList.add("statobject");
                }
                statObject.entry = data;

                labelSpan.textContent = data.label;
                nameSpan.textContent = data.name;
                suffixSpan.textContent = data.suffix;

                const wantedImage = imageLoader.getImage(data.icon);
                if (wantedImage === null) {
                    if (iconDiv.children.length > 0) {
                        iconDiv.children[0].remove();
                    }
                    iconDiv.style.width = "";
                    iconDiv.style.height = "";
                    iconDiv.style.overflow = "";
                } else {
                    if (iconDiv.children.length > 0 && iconDiv.children[0].src !== wantedImage.src) {
                        iconDiv.children[0].remove();
                    }
                    iconDiv.style.width = (data.sw || wantedImage.width) + "px";
                    iconDiv.style.height = (data.sh || wantedImage.height) + "px";
                    iconDiv.style.overflow = "hidden";
                    var img;
                    if (iconDiv.children.length === 0) {
                        img = wantedImage.cloneNode(true);
                        iconDiv.appendChild(img);
                    } else {
                        img = iconDiv.children[0];
                    }
                    img.marginLeft = "-" + (data.sx || 0) + "px";
                    img.marginTop = "-" + (data.sy || 0) + "px";
                }
            }
        }
    }

    function updateStatPanels() {
        var panels = statPanels;
        var allPanels = [];
        var canUseStatPanel = false;
        for (var panel in panels) {
            if (panels.hasOwnProperty(panel)) {
                allPanels.push(panel);
                if (selectedStatPanel !== null && selectedStatPanel === panel) {
                    canUseStatPanel = true;
                }
            }
        }
        allPanels.sort();
        if (verbs.length > 0) {
            allPanels.push("Commands");
            if (selectedStatPanel === "Commands") {
                canUseStatPanel = true;
            }
        }
        if (!canUseStatPanel) {
            selectedStatPanel = allPanels.length > 0 ? allPanels[0] : null;
        }
        renderPanelTabs(allPanels, selectedStatPanel);
        if (selectedStatPanel === "Commands") {
            renderPanelData(verbs, true);
        } else if (selectedStatPanel !== null) {
            var paneldata = panels[selectedStatPanel];
            renderPanelData(paneldata.entries || [], false);
        } else {
            renderPanelData([]);
        }
    }

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
            statPanels = message.newstate.stats.panels || {};
            verbs = message.newstate.verbs || [];
            updateStatPanels();
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
