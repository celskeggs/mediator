# Mediator

The Mediator game engine is designed to function as a usable replacement for the BYOND game engine.

## Disclaimer

This does not work yet. This does not remotely work yet. I'm not even sure it's ever going to work. Don't say I didn't
warn you.

## Overview

It contains the following components:

 * A DreamMaker-inspired object model implemented in Go.
 * A server-side Go game engine, using the object model.
 * A webclient and communication protocol to let users play games in the engine.
 * Parsers for some of BYOND's resource formats, like icons and maps.

It is also intended to contain the following component:

 * A source-to-source transpiler from BYOND to Go.

## Motivation

The intention of this engine is to provide a platform on which complex BYOND games could be allowed to run on free
software platforms: it is intended to be able to be developed and hosted run on Linux, and be played from any modern web
browser on any platform that has a modern web browser.

The intention is not to be perfectly compatible with BYOND; I expect that any large game being ported will require some
patching before and after source-to-source transpilation. That's considered fine, although of course I'd like to
minimize that as much as possible.

(Yes, BYOND has a web client, but my understanding is that it doesn't work well for complex BYOND games.)

This is, of course, a lofty goal, so don't expect this effort to necessarily succeed.

## Trying it out

See the instructions in the [examples repository](https://github.com/celskeggs/mediator-examples/).

## FAQ

#### Why is this not written as a reimplementation of the BYOND compiler and runtime?

Multiple reasons:

 * BYOND is an incredibly complex platform. I don't want to deal with implementing a programming language from scratch if I
can help it. By basing this on Go, and using a source-to-source transpiler, I can avoid having to debug complex machine
code generation issues, and can have an easier time debugging, because the output of the transpiler is at least vaguely
human-readable, unlike machine code.

* Using Go means that I can take advantage of the existing Go ecosystem for things like the HTTP library and
WebSocket support, which makes things way easier.

* It provides a migration path for a game to move out of DreamMaker entirely.

* It means that additional hosting platforms can be more easily supported.

## License

This is currently licensed under the GPL v3, but I'm considering moving it to the AGPL v3 if it gets anywhere.
