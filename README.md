# system-control

A utility to make common system actions a breeze.

## What?
In a linux environment it can be cumbersome to do simple stuff like
* changing the volume
* increasing the display brightness
* manage WiFi connections / Hotspots
because most of them requires knowledge about a specific shell tool.

This project aims to simplify that by providing a command line tool that
can do all of the things mentioned above with a simple and easy to understand syntax.

## Why not use Shell scripts?
Yes, you could probably write scripts for all of the things
`system-control` can do and be happy with that. That was also my initial approach.
When I wrote my first script I tried to use Bash since it is more
close to the system, but:

* Bash scripting syntax is exhausting
* Parsing or extracting text from command output can be very tricky/hacky when compared to a "real" programming language

Because of those downsides I started to use Python scripts.
But python also has its downsides:

* Python is relatively slow (especially when compared to pure bash)
* Dependency management can be tricky when not using something like [Poetry](https://github.com/python-poetry/poetry)
* There are still lots of files to manage

## System

### Shutdown/Restart

```shell script
system-control shutdown
```

```shell script
system-control restart
```

```shell script
system-control lock
```

## Desktop

### Polybar

```shell script
system-control desktop bar start -monitor XXX -bar "alien"
```

## Hardware

### Network

#### Hotspot

```shell script
system-control hotspot on -n "MyHotspot"
```

```shell script
system-control hotspot off -n "MyHotspot"
```

### Screen

#### Brightness

```shell script
system-control display brightness set 100
```

```shell script
system-control display brightness inc
```

```shell script
system-control display brightness set on
```

```shell script
system-control display brightness set off
```

### Audio

#### Volume

```shell script
system-control audio volume
```

```shell script
system-control audio volume inc
system-control audio volume dec
```

```shell script
system-control audio volume set 100 --channel Master
```

```shell script
system-control audio mute
system-control audio unmute
system-control audio toggle-mute
```

#### Sink

```shell script
system-control audio sink
system-control audio sink set "headphone"
system-control audio sink switch "headphone"
```