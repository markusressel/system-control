# system-control

A utility to make common system actions a breeze.

## What?

On linux it can be cumbersome to do simple stuff via the command line, like f. ex.:
* changing the volume
* increasing the display brightness
* manage WiFi connections / Hotspots
* connecting to bluetooth devices

In most cases these actions either require knowledge about a specific shell tool to do the job or even a custom-built script.

This project aims to simplify that by providing a command line tool that can do all the things mentioned above with a
simple and easy to understand syntax.

## Highly opinionated

This project is highly opinionated and only supports the tools I use myself. This is to keep the codebase manageable
and to ensure that the project stays focused on the things I need. If you need support for other tools, feel free to
open an issue or a PR and I will consider adding it. If not, feel free to fork the project and add the tools you need
to your own version of the project.

## Why not use Shell scripts?

Yes, you could write scripts for all the things
`system-control` can do and be happy with that. In fact, that was also my initial approach.
When I wrote my first script I tried to use Bash since it is more close to the system, but:

* Bash scripting syntax is exhausting (imho)
* Parsing or extracting text from command output can be very tricky/hacky when compared to a "real" programming language
* Combining and maintaining multiple commands is tricky since the commands are possibly in separate files

Because of those downsides I started to use Python scripts, but python also has its downsides:

* Python is relatively slow (especially when compared to pure bash) since it has to start a python environment every time
* Dependency management can be tricky when not using something like [Poetry](https://github.com/python-poetry/poetry)
* If you don't want to write a massive one-file script there are still lots of files to manage

To fix the dependency and file management you could use any packaging tool. But since this would
not address the performance concerns I had, I decided to give golang a try. Go allows me to build
a single binary file that includes everything necessary and provides the best possible performance
(without using bash). The performance aspect is crucial for my usage, because some of these commands
get executed on a regular basis through [Polybar](https://github.com/polybar/polybar) and other applications. Of course
it also has its downsides, the main one beeing that it is the least flexible solution since it is a compiled
binary, but history has shown that this is mostly non-issue, because commands on linux tend to have a
very stable API or CLI interface (even when its explicitly stated that its not).

# Installation

Note that you need a working Go environment to build this project.

```shell
> git clone https://github.com/markusressel/system-control.git
> cd system-control
> make deploy
```

If you like, generate shell completions for your shell:

```shell
system-control completion fish > ~/.config/fish/completions/system-control.fish
```

# Usage

system-control is a command-line-interface (CLI) tool. To use it simply open a terminal and type in one
of the listed commands. If you want to see all available commands, type `system-control help`.

```shell
> system-control help
A utility to make common system actions a breeze.

Usage:
  system-control [command]

Available Commands:
  audio       Control System Audio
  battery     Control System Battery
  bluetooth   Control Bluetooth Devices
  completion  Generate the autocompletion script for the specified shell
  display     Control Display
  help        Help about any command
  keyboard    Control Keyboard
  restart     Reboot the system gracefully
  shutdown    Shutdown the system gracefully
  touchpad    Control touchpad
  video       Control video inputs (cameras)

Flags:
      --configuration string   configuration file (default is $HOME/.system-control.yaml)
  -h, --help                   help for system-control
  -t, --toggle                 Help message for toggle

Use "system-control [command] --help" for more information about a command.
```

## Audio

system-control is optimized for pipewire. If currently you are not using pipewire already, I strongly recommend to
consider it.

**Requirements:**

* `pw-dump`
* `pw-cli`
* `wpctl`

```shell
> system-control audio volume mute
> system-control audio volume unmute
> system-control audio volume toggle-mute
```

```shell
> system-control audio volume
28
> system-control audio volume inc
> system-control audio volume dec
> system-control audio volume set 100 --channel Master
```

Save and Restore Audio State, f.ex. before and after reboot:

```shell
> system-control audio volume save
> system-control audio volume restore
```

### Sink

```shell
// list sinks
> system-control audio sink
Sink #64
	State: SUSPENDED
	Name: alsa_output.pci-0000_11_00.4.analog-stereo
	Description: Starship/Matisse HD Audio Controller Analog Stereo
	Driver: PipeWire
	Sample Specification: s32le 2ch 48000Hz
	Channel Map: front-left,front-right
[...]

// get active sink
> system-control audio sink active
46

// check if the current active sink contains "nvidia" in its name
> system-control audio sink active "NVIDIA"
0

// switch active sink to a sink which contains the given text
> system-control audio sink switch "Built-in"
> system-control audio sink switch "X-Fi"
> system-control audio sink switch "NVIDIA"

> system-control audio sink next
> system-control audio sink previous
```

### Battery

```shell
> system-control battery list
hidpp_battery_0
  Path:         /sys/class/power_supply/hidpp_battery_0
  Type:         Battery
  Manufacturer: Logitech
  Model:        MX Keys Wireless Keyboard
  Serial:       a8-bc-3a-a5
  Capacity:     Full
  Online:       true
  Status:       Discharging
  Scope:        Device

> system-comtrol battery remaining

> system-control battery threshold -name BAT0
100
> system-control battery threshold -name BAT0 75

> system-control battery threshold -name BAT0 save       # run this after changing the threshold
> system-control battery threshold -name BAT0 restore    # run this f.ex. right after boot
```

### Bluetooth

```shell
> system-control bluetooth on
> system-control bluetooth off
```

```shell
> system-control bluetooth devices
```

```shell
> system-control bluetooth pair "LG-TONE-FP9"
> system-control bluetooth pair "00:1D:43:6D:03:1A"
```

```shell
> system-control bluetooth remove "LG-TONE-FP9"
> system-control bluetooth remove "00:1D:43:6D:03:1A"
```

```shell
> system-control bluetooth connect "LG-TONE-FP9"
> system-control bluetooth connect "00:1D:43:6D:03:1A"
```

```shell
> system-control bluetooth disconnect "LG-TONE-FP9"
> system-control bluetooth disconnect "00:1D:43:6D:03:1A"

# disconnect all devices
> system-control bluetooth disconnect
```

### Display / Screen

#### List Screens

```shell
> system-control display list
DisplayPort-2
DisplayPort-1
````

#### Backlight

```shell
> system-control display backlight list

> system-control display backlight brightness set 100
> system-control display backlight brightness inc
> system-control display backlight brightness dec
```

#### RedShift

```shell
> system-control display redshift
Display: DisplayPort-2
  Color Temperature: 4500
  Brightness: 0.70
  Gamma: -1.00
Display: DisplayPort-1
  Color Temperature: 4500
  Brightness: 0.70
  Gamma: -1.00
```

```shell
> system-control display redshift reset
> system-control display redshift update
```

## Keyboard

### Brightness

```shell
> system-control keyboard brightness set 100
> system-control keyboard brightness inc
> system-control keyboard brightness dec
```

## Touchpad

```shell
> system-control touchpad on
> system-control touchpad off
> system-control touchpad set true
> system-control touchpad toggle
```

## Video (Webcam)

```shell
> system-control video load    # loads the kernel driver for webcam support
> system-control video unload  # unloads the kernel driver for webcam support
```

### Network

#### Hotspot

TODO: not yet implemented

```shell
> system-control hotspot on -n "MyHotspot"
```

```shell
> system-control hotspot off -n "MyHotspot"
```

### Shutdown/Restart

TODO: not yet fully working

```shell
> system-control shutdown
```

```shell
> system-control restart
```

```shell
> system-control lock
```