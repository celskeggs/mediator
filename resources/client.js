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

function startSoundPlayer() {
    var channels = {};
    var maxChannel = 1024;
    var Player = {};

    function validateSound(sound) {
        return sound.channel !== undefined && sound.channel !== null && sound.channel <= maxChannel;
    }

    function getChannel(channel) {
        if (channel < 1 || channel > 1024) {
            return null;
        }
        if (channels[channel] === undefined) {
            channels[channel] = {
                "element": null,
                "current": null,
                "queue": []
            };
        }
        return channels[channel];
    }

    function findAvailableChannel() {
        for (var i = 1; i <= maxChannel; i++) {
            if (getChannel(i).current === null) {
                return i;
            }
        }
        return 0;
    }

    function cancelSounds(channelId) {
        var channel = getChannel(channelId);
        if (channel.element !== null) {
            channel.element.pause();
        }
        channel.current = null;
        channel.queue.splice(0);
    }

    function startPlaying(channelId) {
        var channel = getChannel(channelId);
        console.log("playing", channel.current.file, "at volume", channel.current.volume);
        channel.element.src = "resource/" + channel.current.file;
        channel.element.volume = channel.current.volume / 100.0;
        channel.element.play();
    }

    function onEnd(channelId) {
        var channel = getChannel(channelId);
        if (channel.current === null) {
            console.log("ended when nothing was supposed to be playing");
            return;
        }
        if (channel.current.repeat) {
            console.log("continuing loop");
            channel.element.play();
            return;
        }
        if (channel.queue.length > 0) {
            channel.current = channel.queue.shift();
            startPlaying(channelId);
        }
    }

    function queueSound(channelId, sound) {
        var channel = getChannel(channelId);
        if (channel.element === null) {
            channel.element = new Audio();
            channel.element.addEventListener("ended", function (ev) {
                onEnd(channelId);
            });
        }
        if (channel.current === null) {
            channel.current = sound;
            startPlaying(channelId);
        } else {
            channel.queue.push(sound);
        }
    }

    Player.playSound = function (sound) {
        if (!validateSound(sound)) {
            console.log("invalid sound", sound);
            return;
        }
        if (sound.file === null) {
            if (sound.channel === 0) {
                // TODO: affect all channels with these settings
            } else {
                console.log("ignoring sound without file", sound);
            }
            return;
        }
        var channel = sound.channel;
        if (channel <= 0) {
            // choose any available channel
            channel = findAvailableChannel();
            if (channel <= 0) {
                console.log("could not find channel for sound; ignoring");
                return;
            }
        }
        if (sound.file) {
            if (!sound.wait) {
                cancelSounds(channel);
            }
            queueSound(channel, sound);
        } else {
            // TODO: is this the right behavior?
            cancelSounds(channel);
        }
    };

    Player.cancelAllSounds = function () {
        for (var i = 1; i <= maxChannel; i++) {
            var channel = channels[i];
            if (channel !== undefined) {
                cancelSounds(i);
            }
        }
    };

    return Player;
}

function prepareGame(canvas, inputsource, verbentry, paneltabs, panelbody, textoutput) {
    var images = null;
    var isTerminated = false;
    var gameActive = false;
    var width = 672, height = 672;
    var aspectRatio = width / height;
    var aspectShiftX = 0, aspectShiftY = 0;
    var gameSprites = [];
    var statPanels = {};
    var verbs = [];
    var selectedStatPanel = null;
    var keyDirection = null;
    var sendMessage = function (message) {
    };
    var Player = startSoundPlayer();

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
                child.appendChild(document.createElement("div"));
                child.appendChild(document.createElement("span"));
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
                var iconDiv = entry.children[1];
                var nameSpan = entry.children[2];
                var suffixSpan = entry.children[3];

                labelSpan.textContent = data.label;
                nameSpan.textContent = data.name;
                suffixSpan.textContent = data.suffix;

                if (images !== null) {
                    var wantedImage = null;
                    if (data.icon !== "" && data.icon in images) {
                        wantedImage = images[data.icon];
                    }
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
        console.log("selected", selectedStatPanel);
        if (selectedStatPanel === "Commands") {
            renderPanelData(verbs, true);
        } else if (selectedStatPanel !== null) {
            var paneldata = panels[selectedStatPanel];
            renderPanelData(paneldata.entries || [], false);
        } else {
            renderPanelData([]);
        }
    }

    function onMessage(message) {
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
            console.log("state", message.newstate);
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
    }

    function onConnectionClosed() {
        Player.cancelAllSounds();
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

    verbentry.focus();

    verbentry.addEventListener("keypress", function (ev) {
        if (ev.key === "Enter" && verbentry.value !== "") {
            sendVerb(verbentry.value);
            verbentry.value = "";
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
    var verbEntry = document.getElementById("verb");
    var panelTabs = document.getElementById("paneltabs");
    var panelBody = document.getElementById("panelbody");
    var textOutput = document.getElementById("textspace");
    if (canvas.getContext) {
        prepareGame(canvas, document.body, verbEntry, panelTabs, panelBody, textOutput);
    }
});
