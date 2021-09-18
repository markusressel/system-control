/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package cmd

import (
	"bytes"
	. "fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	DisplayBacklightPath = "/sys/class/backlight"
	MaxBrightness        = "max_brightness"
	Brightness           = "brightness"
)

func isMuted(channel string) bool {
	result, err := execCommand("amixer", "-D", "pulse", "get", channel)
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("\\[(on|off)]")
	match := r.FindString(result)
	return match == "[off]"
}

func setMuted(channel string, muted bool) {
	var state string
	if muted {
		state = "off"
	} else {
		state = "on"
	}

	_, err := execCommand("amixer", "-D", "pulse", "set", channel, state)
	if err != nil {
		log.Fatal(err)
	}
}

func getVolume(channel string) int {
	result, err := execCommand("amixer", "-D", "pulse", "get", channel)
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

// Calculates an appropriate amount of volume change when the user did not specify a specific value.
func calculateAppropriateVolumeChange(current int, increase bool) int {
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

func setVolume(channel string, volume int) {
	_, err := execCommand("amixer", "-D", "pulse", "set", channel, strconv.Itoa(volume)+"%")
	if err != nil {
		log.Fatal(err)
	}
}

// returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func findActiveSinkPulse(text string) int {
	// ignore case
	text = strings.ToLower(text)

	result, err := execCommand("pactl", "list", "sinks")
	if err != nil {
		log.Fatal(err)
	}

	// we dont need case information
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

// returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func findActiveSinkPipewire(text string) int {
	// ignore case
	text = strings.ToLower(text)

	currentDefaultSinkName, err := execCommand("pactl", "get-default-sink")
	if err != nil {
		log.Fatal(err)
	}

	var activeSinkIndex int
	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if strings.Contains(strings.ToLower(item["node.name"]), currentDefaultSinkName) {
			activeSinkIndex, err = strconv.Atoi(item["id"])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if len(text) > 0 {
		sinkIndex := findSinkPipewire(text)
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

	result, err := execCommand("pactl", "list", "sinks")
	if err != nil {
		log.Fatal(err)
	}
	// we dont need case information
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

// returns the index of a sink that contains the given text
func findSinkPipewire(text string) int {
	// ignore case
	text = strings.ToLower(text)

	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if strings.Contains(strings.ToLower(item["node.description"]), text) {
			index, err := strconv.Atoi(item["id"])
			if err != nil {
				log.Fatal(err)
			}
			return index
		}
	}

	return -1
}

type PropertyFilter struct {
	key   string
	value string
}

// retrieve a list of pipewire objects
// optionally filtered
func getPipewireObjects(filters ...PropertyFilter) (objects []map[string]string) {
	result, err := execCommand("pw-cli", "ls")
	if err != nil {
		log.Fatal(err)
	}

	objects = parsePipwireToMap(result)
	objects = FilterPipwireObjects(objects, func(v map[string]string) bool {
		for _, filter := range filters {
			if v[filter.key] != filter.value {
				return false
			}
		}

		return true
	})

	return objects
}

func parsePipwireToMap(input string) []map[string]string {
	var result = make([]map[string]string, 0, 1000)

	lines := strings.Split(input, "\n")
	var objectMap map[string]string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) <= 0 {
			continue
		}
		if strings.Contains(line, ",") && !strings.Contains(line, "=") {
			// this is the "header" of an object

			// create a new map for the current object and fill it
			objectMap = make(map[string]string)
			splits := strings.Split(line, ",")
			for _, item := range splits {
				item = strings.TrimSpace(item)
				splits := strings.SplitAfter(item, " ")
				objectMap[strings.TrimSpace(splits[0])] = strings.TrimSpace(splits[1])
			}
			result = append(result, objectMap)
		} else {
			// this is a property of an object

			splits := strings.SplitAfter(line, "=")

			key := strings.TrimRight(splits[0], "=")
			key = strings.TrimSpace(key)

			value := strings.TrimSpace(splits[1])
			value = strings.Trim(value, "\"")
			objectMap[key] = value
		}
	}

	return result
}

// FilterPipwireObjects filters the given pipewire object map based on the given function
func FilterPipwireObjects(vs []map[string]string, f func(map[string]string) bool) []map[string]string {
	vsf := make([]map[string]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Switches the default sink to the target sink
func setDefaultSinkPulse(index int) (err error) {
	indexString := strconv.Itoa(index)
	_, err = execCommand("pactl", "set-default-sink", indexString)
	return err
}

// Switches the default sink to the target sink
// You need to get a sink name with "pw-cli ls Node"
// and look for the "node.name" property for a valid value.
func setDefaultSinkPipewire(sinkName string) (err error) {
	// TODO: not sure if this switches running apps over
	_, err = execCommand("pw-metadata", "0", "default.configured.audio.sink", `'{ "name": "`+sinkName+`" }'`)
	return err
}

// Switches the default sink and moves all existing sink inputs to the target sink
func switchSinkPulse(index int) {
	err := setDefaultSinkPulse(index)
	if err != nil {
		log.Fatal(err)
	}

	indexString := strconv.Itoa(index)
	result, err := execCommand("pactl", "list", "sink-inputs", indexString)
	if err != nil {
		log.Fatal(err)
	}

	ri := regexp.MustCompile("index: (\\d+)")
	matches := ri.FindAllStringSubmatch(result, -1)

	for i := range matches {
		inputIdx := matches[i][1]
		_, err := execCommand("pactl", "move-sink-input", inputIdx, indexString)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Switches the default sink and moves all existing sink inputs to the target sink
func switchSinkPipewire(index int) {
	err := setDefaultSinkPulse(index)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: find all applications
	// TOOD: find outputs of applications
	// TODO: find current links of applications
	// TODO: find inputs of target sink
	// TODO: modify/delete and create links for all applications to the new target
}

func readIntFromFile(path string) (int64, error) {
	fileBuffer, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	value := string(fileBuffer)
	value = strings.TrimSpace(value)
	return strconv.ParseInt(value, 0, 64)
}

func writeIntToFile(value int, path string) error {
	fileStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	return ioutil.WriteFile(path, []byte(strconv.Itoa(value)), fileStat.Mode())
}

func getMaxBrightness() int {
	backlightName := findBacklight()
	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	maxBrightness, err := readIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(maxBrightness)
}

func getBrightness() int {
	backlightName := findBacklight()
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := readIntFromFile(brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(brightness)
}

// Sets a specific brightness of main the display
func setBrightness(percentage int) {
	files, err := ioutil.ReadDir(DisplayBacklightPath)
	if err != nil {
		log.Fatal(err)
	}

	var backlightName string
	if len(files) <= 1 {
		backlightName = files[0].Name()
	} else {
		// TODO: select first? select by user input?
	}

	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness

	maxBrightness, err := readIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	targetValue := int((float32(percentage) / 100.0) * float32(maxBrightness))
	err = writeIntToFile(targetValue, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	//env := []string{"DISPLAY:=0"}
	//command := "-set"
	//_, err := execCommandEnv(env, true, "xbacklight", command, strconv.Itoa(percentage), "-steps", "1", "-time", "0")
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func setBrightnessRaw(backlight string, brightness int) {
	maxBrightness := getMaxBrightness()
	targetBrightness := brightness
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlight + string(os.PathSeparator) + Brightness

	err := writeIntToFile(targetBrightness, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
}

// Adjusts the brightness of the main display
func adjustBrightness(change int) {
	backlight := findBacklight()

	maxBrightness := getMaxBrightness()
	currentBrightness := getBrightness()

	targetBrightness := currentBrightness + change
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	setBrightnessRaw(backlight, targetBrightness)
}

func findBacklight() string {
	files, err := ioutil.ReadDir(DisplayBacklightPath)
	if err != nil {
		log.Fatal(err)
	}

	var backlightName string
	if len(files) <= 1 {
		backlightName = files[0].Name()
	} else {
		// TODO: select first? select by user input?
	}

	return backlightName
}

func findOpenWindows() []string {
	result, err := execCommand("wmctrl", "-l")
	if err != nil {
		log.Fatal(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("(0x[0-9a-f]+) +(\\d+) +(" + hostname + "|N/A) +(.*)")
	matches := r.FindAllString(result, -1)
	return matches
}

// Executes a shell command with the given arguments
// and returns its stdout as a []byte.
// If an error occurs the content of stderr is printed
// and an error is returned.
func execCommand(command string, args ...string) (string, error) {
	//log.Printf("Executing command: %s %s", command, args)
	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		Println(err.Error())
		Println(string(stderr.Bytes()))
		return "", err
	}

	result := string(stdout.Bytes())
	result = strings.TrimSpace(result)

	return result, nil
}

// Like execCommand but with the possibility to add environment variables
// to the executed process.
func execCommandEnv(env []string, attach bool, command string, args ...string) (string, error) {
	//log.Printf("Executing command: %s %s", command, args)
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	var err error
	if attach {
		err = cmd.Run()
	} else {
		err = cmd.Start()
		if err != nil {
			Println(err.Error())
			return "", err
		}
		err = cmd.Process.Release()
	}

	if err != nil {
		Println(err.Error())
		Println(string(stderr.Bytes()))
		log.Fatal(stderr)
		return "", err
	}

	return string(stdout.Bytes()), nil
}
