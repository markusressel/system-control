package pulseaudio

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

func IsMuted(card int, channel string) bool {
	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "get", channel)

	result, err := util.ExecCommand("amixer", args...)
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("\\[(on|off)]")
	match := r.FindString(result)
	return match == "[off]"
}

func SetMuted(card int, channel string, muted bool) error {
	var state string
	if muted {
		state = "off"
	} else {
		state = "on"
	}

	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "set", channel, state)

	_, err := util.ExecCommand("amixer", args...)
	return err
}

func GetVolume(card int, channel string) int {
	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "get", channel)

	result, err := util.ExecCommand("amixer", args...)
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("\\[\\d+%]")
	match := r.FindString(result)
	match = match[1 : len(match)-2]
	volume, err := strconv.Atoi(match)
	if err != nil {
		log.Fatal(err)
	}
	return volume
}

func SetVolume(card int, channel string, volume int) error {
	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "set", channel, strconv.Itoa(volume)+"%")

	_, err := util.ExecCommand("amixer", args...)
	return err
}

// Switches the default sink to the target sink
func setDefaultSinkPulse(index int) (err error) {
	indexString := strconv.Itoa(index)
	_, err = util.ExecCommand("pactl", "set-default-sink", indexString)
	return err
}

// Switches the default sink and moves all existing sink inputs to the target sink
func switchSinkPulse(index int) {
	err := setDefaultSinkPulse(index)
	if err != nil {
		log.Fatal(err)
	}

	indexString := strconv.Itoa(index)
	result, err := util.ExecCommand("pactl", "list", "sink-inputs", indexString)
	if err != nil {
		log.Fatal(err)
	}

	ri := regexp.MustCompile("index: (\\d+)")
	matches := ri.FindAllStringSubmatch(result, -1)

	for i := range matches {
		inputIdx := matches[i][1]
		_, err := util.ExecCommand("pactl", "move-sink-input", inputIdx, indexString)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func findActiveSinkPulse(text string) int {
	// ignore case
	text = strings.ToLower(text)

	result, err := util.ExecCommand("pactl", "list", "sinks")
	if err != nil {
		log.Fatal(err)
	}

	// we don't need case information
	result = strings.ToLower(result)

	// search for the Index line containing a star
	ri := regexp.MustCompile("(?i)\\s+\\*\\s+Index: \\d+")
	matches := ri.FindAllString(result, -1)
	match := matches[len(matches)-1]

	// extract the activeSinkIndex number
	rd := regexp.MustCompile("(?i)\\d+")
	activeSinkIndexMatch := rd.FindString(match)
	activeSinkIndex, err := strconv.Atoi(activeSinkIndexMatch)
	if err != nil {
		log.Fatal(err)
	}

	if len(text) > 0 {
		sinkIndex := findSinkPulse(text)
		if sinkIndex == activeSinkIndex {
			return 1
		} else {
			return 0
		}
	} else {
		return activeSinkIndex
	}
}

// returns the index of a sink that contains the given text
func findSinkPulse(text string) int {
	// ignore case
	text = strings.ToLower(text)

	result, err := util.ExecCommand("pactl", "list", "sinks")
	if err != nil {
		log.Fatal(err)
	}
	// we don't need case information
	result = strings.ToLower(result)

	// find the wanted text
	i := strings.Index(result, text)
	if i == -1 {
		log.Fatalf("Substring %s not found", text)
	}

	substring := result[0 : i+len(text)]
	// search bottom-up for the first "index" line before the matched text line
	ri := regexp.MustCompile("(?i)index: \\d+")
	matches := ri.FindAllString(substring, -1)
	match := matches[len(matches)-1]

	// extract the index number
	rd := regexp.MustCompile("(?i)\\d+")
	sinkIndex := rd.FindString(match)
	index, err := strconv.Atoi(sinkIndex)
	if err != nil {
		log.Fatal(err)
	}

	return index
}

// SetVolumePulseAudio sets the given volume to the given sink using PulseAudio
// volume in percent
func SetVolumePulseAudio(sinkId int, volume float64) error {
	_, err := util.ExecCommand(
		"pactl",
		"set-sink-volume",
		strconv.Itoa(sinkId),
		fmt.Sprint(volume),
	)
	return err
}
