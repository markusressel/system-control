# system-control

A utility to make common system actions a breeze.

## What?
On linux it can be cumbersome to do simple stuff via the command line, like f. ex.:
* changing the volume
* increasing the display brightness
* manage WiFi connections / Hotspots
* graceful shutting/restart
In most cases these actions either require knowledge about a specific shell tool
to do the job or even a custom-built script.

This project aims to simplify that by providing a command line tool that
can do all the things mentioned above with a simple and easy to understand syntax.

## Why not use Shell scripts?
Yes, you could write scripts for all the things
`system-control` can do and be happy with that. That was also my initial approach.
When I wrote my first script I tried to use Bash since it is more
close to the system, but:

* Bash scripting syntax is exhausting
* Parsing or extracting text from command output can be very tricky/hacky when compared to a "real" programming language

Because of those downsides I started to use Python scripts, but python also has its downsides:

* Python is relatively slow (especially when compared to pure bash) since it has to start a python environment every time
* Dependency management can be tricky when not using something like [Poetry](https://github.com/python-poetry/poetry)
* If you don't want to write a massive one-file script there are still lots of files to manage

To fix the dependency and file management you could use any packaging tool. But since this would
not address the performance concerns I had, I decided to give golang a try. Go allows me to build
a single binary file that includes everything necessary and provides the best possible performance
(without using bash). The performance aspect is crucial for my usage, because some of these commands
 get executed on a regular basis through [Polybar](https://github.com/polybar/polybar) and other 
 applications.

## System

### Shutdown/Restart

TODO:

```shell script
system-control shutdown
```

```shell script
system-control restart
```

```shell script
system-control lock
```

## Hardware

### Network

#### Hotspot

TODO: not yet implemented

```shell script
system-control hotspot on -n "MyHotspot"
```

```shell script
system-control hotspot off -n "MyHotspot"
```

### Screen

#### Brightness

**Requirements:**
* `xbacklight` (soon to be removed)

```shell script
system-control display brightness set 100
system-control display brightness inc
system-control display brightness dec
```

### Audio

**Requirements:**
* `pacmd`
* `amixer`

#### Volume

```shell script
system-control audio mute
system-control audio unmute
system-control audio toggle-mute
```

```shell script
system-control audio volume
system-control audio volume inc
system-control audio volume dec
system-control audio volume set 100 --channel Master
```

#### Sink

```shell script
// list sinks
system-control audio sink

// get active sink
system-control audio sink active
system-control audio sink active "NVIDIA"

// switch active sink
system-control audio sink switch "Built-in"
system-control audio sink switch "X-Fi"
system-control audio sink switch "NVIDIA"
```