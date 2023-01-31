package audio

import (
	"github.com/markusressel/system-control/internal/util"
	"log"
	"regexp"
	"strconv"
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

func IsHeadphoneConnected() bool {
	// TODO:
	return false
}

// Calculates an appropriate amount of volume change when the user did not specify a specific value.
func CalculateAppropriateVolumeChange(current int, increase bool) int {
	localCurrent := current

	if !increase {
		localCurrent--
	}

	if localCurrent < 20 {
		return 1
	} else if localCurrent < 40 {
		return 2
	} else {
		return 5
	}
}
