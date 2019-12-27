"use strict";

/* SoundPlayer is a helper object that manages sounds currently being played.
 * The public interface is the following two methods:
 *  - playSound(soundobject)
 *  - cancelAllSounds()
 */

function SoundPlayer() {
    this.channels = {};
    this.maxChannel = 1024;
}

SoundPlayer.prototype.validateSound = function (sound) {
    return sound.channel !== undefined && sound.channel !== null && sound.channel <= this.maxChannel;
};

SoundPlayer.prototype.getChannel = function (channel) {
    if (channel < 1 || channel > 1024) {
        return null;
    }
    if (this.channels[channel] === undefined) {
        this.channels[channel] = {
            "element": null,
            "current": null,
            "queue": []
        };
    }
    return this.channels[channel];
};

SoundPlayer.prototype.findAvailableChannel = function () {
    for (let i = 1; i <= this.maxChannel; i++) {
        if (this.getChannel(i).current === null) {
            return i;
        }
    }
    return 0;
};

SoundPlayer.prototype.cancelSounds = function (channelId) {
    const channel = this.getChannel(channelId);
    if (channel.element !== null) {
        channel.element.pause();
    }
    channel.current = null;
    channel.queue.splice(0);
};

SoundPlayer.prototype.startPlaying = function (channelId) {
    const channel = this.getChannel(channelId);
    console.log("playing", channel.current.file, "at volume", channel.current.volume);
    channel.element.src = "resource/" + channel.current.file;
    channel.element.volume = channel.current.volume / 100.0;
    channel.element.play();
};

SoundPlayer.prototype.onEnd = function (channelId) {
    const channel = this.getChannel(channelId);
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
        this.startPlaying(channelId);
    }
};

SoundPlayer.prototype.queueSound = function (channelId, sound) {
    const outerThis = this;
    const channel = this.getChannel(channelId);
    if (channel.element === null) {
        channel.element = new Audio();
        channel.element.addEventListener("ended", function () {
            outerThis.onEnd(channelId);
        });
    }
    if (channel.current === null) {
        channel.current = sound;
        this.startPlaying(channelId);
    } else {
        channel.queue.push(sound);
    }
};

SoundPlayer.prototype.playSound = function (sound) {
    if (!this.validateSound(sound)) {
        console.log("invalid sound", sound);
        return;
    }
    if (sound.file === null) {
        if (sound.channel === 0) {
            // FIXME: affect all channels with these settings
        } else {
            console.log("ignoring sound without file", sound);
        }
        return;
    }
    let channel = sound.channel;
    if (channel <= 0) {
        // choose any available channel
        channel = this.findAvailableChannel();
        if (channel <= 0) {
            console.log("could not find channel for sound; ignoring");
            return;
        }
    }
    if (sound.file) {
        if (!sound.wait) {
            this.cancelSounds(channel);
        }
        this.queueSound(channel, sound);
    } else {
        // FIXME: determine whether this is the right behavior
        this.cancelSounds(channel);
    }
};

SoundPlayer.prototype.cancelAllSounds = function () {
    for (let i = 1; i <= this.maxChannel; i++) {
        var channel = this.channels[i];
        if (channel !== undefined) {
            this.cancelSounds(i);
        }
    }
};
