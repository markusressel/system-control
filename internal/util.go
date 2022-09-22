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
package internal

import (
	"bytes"
	. "fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	DisplayBacklightPath = "/sys/class/backlight"
	LedsPath             = "/sys/class/leds"
	MaxBrightness        = "max_brightness"
	Brightness           = "brightness"
)

func IsMuted(card int, channel string) bool {
	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "get", channel)

	result, err := ExecCommand("amixer", args...)
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

	_, err := ExecCommand("amixer", args...)
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

	result, err := ExecCommand("amixer", args...)
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

func SetVolume(card int, channel string, volume int) error {
	var args []string
	if card >= 0 {
		args = append(args, "-c", strconv.Itoa(card))
	} else {
		args = append(args, "-D", "pulse")
	}
	args = append(args, "set", channel, strconv.Itoa(volume)+"%")

	_, err := ExecCommand("amixer", args...)
	return err
}

func IsHeadphoneConnected() bool {
	// TODO:
	return false
}

// returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func findActiveSinkPulse(text string) int {
	// ignore case
	text = strings.ToLower(text)

	result, err := ExecCommand("pactl", "list", "sinks")
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

// returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func FindActiveSinkPipewire(text string) int {
	// ignore case
	text = strings.ToLower(text)

	currentDefaultSinkName, err := ExecCommand("pactl", "get-default-sink")
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
		sink := FindSinkPipewire(text)
		sinkIndex, err := strconv.Atoi(sink["id"])
		if err != nil {
			log.Fatal(err)
		}
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

	result, err := ExecCommand("pactl", "list", "sinks")
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

// FindSinkPipewire returns the index of a sink that contains the given text
func FindSinkPipewire(text string) map[string]string {
	// ignore case
	text = strings.ToLower(text)

	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if strings.Contains(strings.ToLower(item["node.description"]), text) {
			return item
		}
	}

	return nil
}

type PropertyFilter struct {
	key   string
	value string
}

// retrieve a list of pipewire objects
// optionally filtered
func getPipewireObjects(filters ...PropertyFilter) (objects []map[string]string) {
	result, err := ExecCommand("pw-cli", "ls")
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
	_, err = ExecCommand("pactl", "set-default-sink", indexString)
	return err
}

// Switches the default sink to the target sink
// You need to get a sink name with "pw-cli ls Node"
// and look for the "node.name" property for a valid value.
func setDefaultSinkPipewire(sinkName string) (err error) {
	_, err = ExecCommand("pw-metadata", "0", "default.configured.audio.sink", `{ "name": "`+sinkName+`" }`)
	return err
}

// Switches the default sink and moves all existing sink inputs to the target sink
func switchSinkPulse(index int) {
	err := setDefaultSinkPulse(index)
	if err != nil {
		log.Fatal(err)
	}

	indexString := strconv.Itoa(index)
	result, err := ExecCommand("pactl", "list", "sink-inputs", indexString)
	if err != nil {
		log.Fatal(err)
	}

	ri := regexp.MustCompile("index: (\\d+)")
	matches := ri.FindAllStringSubmatch(result, -1)

	for i := range matches {
		inputIdx := matches[i][1]
		_, err := ExecCommand("pactl", "move-sink-input", inputIdx, indexString)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// SwitchSinkPipewire switches the default sink and moves all existing sink inputs to the target sink
func SwitchSinkPipewire(node map[string]string) {
	nodeName := node["node.name"]
	nodeId, err := strconv.Atoi(node["id"])
	if err != nil {
		log.Fatal(err)
	}
	err = setDefaultSinkPipewire(nodeName)
	if err != nil {
		log.Fatal(err)
	}

	var streams = getPipewireObjects(
		PropertyFilter{"media.class", "Stream/Output/Audio"},
	)
	for _, stream := range streams {
		moveStreamToNode(stream["id"], nodeId)
	}
}

func moveStreamToNode(streamId string, nodeId int) {
	_, err := ExecCommand("pw-metadata", streamId, "target.node", strconv.Itoa(nodeId))
	if err != nil {
		log.Fatal(err)
	}
}

func ReadIntFromFile(path string) (int64, error) {
	fileBuffer, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	value := string(fileBuffer)
	value = strings.TrimSpace(value)
	return strconv.ParseInt(value, 0, 64)
}

func WriteIntToFile(value int, path string) error {
	touch(path)
	fileStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return os.WriteFile(path, []byte(strconv.Itoa(value)), fileStat.Mode())
}

func touch(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else if err != nil {
		panic(err)
	}
}

func GetMaxBrightness() int {
	backlightName := findBacklight()
	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(maxBrightness)
}

func GetBrightness() int {
	backlightName := findBacklight()
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(brightness)
}

// SetBrightness sets a specific brightness of main the display
func SetBrightness(percentage int) {
	backlightName := findBacklight()
	maxBrightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness
	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	targetValue := int((float32(percentage) / 100.0) * float32(maxBrightness))
	err = WriteIntToFile(targetValue, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
}

func setBrightnessRaw(backlight string, brightness int) {
	maxBrightness := GetMaxBrightness()
	targetBrightness := brightness
	if targetBrightness < 0 {
		targetBrightness = 0
	}
	if targetBrightness > maxBrightness {
		targetBrightness = maxBrightness
	}

	brightnessPath := DisplayBacklightPath + string(os.PathSeparator) + backlight + string(os.PathSeparator) + Brightness

	err := WriteIntToFile(targetBrightness, brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
}

// AdjustBrightness adjusts the brightness of the main display
func AdjustBrightness(change int) {
	backlight := findBacklight()

	maxBrightness := GetMaxBrightness()
	currentBrightness := GetBrightness()

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
	files, err := os.ReadDir(DisplayBacklightPath)
	if err != nil {
		log.Fatal(err)
	}

	var backlightName string
	if len(files) == 0 {
		log.Fatal("No backlight found")
	} else if len(files) == 1 {
		backlightName = files[0].Name()
	} else {
		// TODO: select first? select by user input?
		backlightName = files[0].Name()
		log.Printf("Found multiple backlight sources, using: " + backlightName)
	}

	return backlightName
}

func findKeyboardBacklight() string {
	files, err := os.ReadDir(LedsPath)
	if err != nil {
		log.Fatal(err)
	}

	var kbdBacklight string
	r := regexp.MustCompile(".*(kbd|keyboard).*")
	for _, f := range files {
		if r.MatchString(f.Name()) {
			return f.Name()
		}
	}

	log.Fatal("No keyboard backlight found")
	return kbdBacklight
}

func GetKeyboardBrightness() int {
	backlightName := findKeyboardBacklight()
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	brightness, err := ReadIntFromFile(brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(brightness)
}

func SetKeyboardBrightness(brightness int) int {
	backlightName := findKeyboardBacklight()
	brightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + Brightness
	maxBrightnessPath := LedsPath + string(os.PathSeparator) + backlightName + string(os.PathSeparator) + MaxBrightness

	maxBrightness, err := ReadIntFromFile(maxBrightnessPath)
	if err != nil {
		log.Fatal(err)
	}

	targetValue := math.Max(0, math.Min(float64(maxBrightness), float64(brightness)))
	err = WriteIntToFile(int(targetValue), brightnessPath)
	if err != nil {
		log.Fatal(err)
	}
	return int(targetValue)
}

func GetInputDevices() []string {
	result, _ := ExecCommand("xinput", "list", "--name-only")
	return strings.Split(result, "\n")
}

func IsInputDeviceEnabled(name string) bool {
	result, _ := ExecCommand("xinput", "list", name)
	return !strings.Contains(result, "This device is disabled")
}

func EnableInputDevice(name string) {
	_, err := ExecCommand("xinput", "enable", name)
	if err != nil {
		log.Fatal(err)
	}
}

func DisableInputDevice(name string) {
	_, err := ExecCommand("xinput", "disable", name)
	if err != nil {
		log.Fatal(err)
	}
}

func GetTouchpadInputDevice() *string {
	inputDevices := GetInputDevices()
	for _, device := range inputDevices {
		if strings.Contains(device, "Touchpad") {
			return &device
		}
	}

	return nil
}

func IsTouchpadEnabledLibinput() bool {
	touchpadDevice := GetTouchpadInputDevice()
	if touchpadDevice != nil {
		return IsInputDeviceEnabled(*touchpadDevice)
	} else {
		return false
	}
}

func IsTouchpadEnabledSynaptics() bool {
	result, _ := ExecCommand("synclient")
	regex := regexp.MustCompile("\\s*TouchpadOff\\s*=\\s*(\\d)")

	submatch := regex.FindStringSubmatch(result)[0]
	submatch = strings.TrimSpace(submatch)
	value := submatch[len(submatch)-1:]

	resultInt, _ := strconv.Atoi(value)
	return resultInt == 0
}

func IsTouchpadEnabled() bool {
	return IsTouchpadEnabledSynaptics() && IsTouchpadEnabledLibinput()
}

func SetTouchpadEnabled(enabled bool) {
	SetTouchpadEnabledSynaptics(enabled)
	SetTouchpadEnabledLibinput(enabled)
}

func SetTouchpadEnabledSynaptics(enabled bool) {
	var enabledInt int
	if enabled {
		enabledInt = 0
	} else {
		enabledInt = 1
	}

	_, err := ExecCommand("synclient", "TouchpadOff="+strconv.Itoa(enabledInt))
	if err != nil {
		log.Fatal(err)
	}
}

func SetTouchpadEnabledLibinput(enabled bool) {
	touchpadDevice := GetTouchpadInputDevice()
	if touchpadDevice != nil {
		if enabled {
			EnableInputDevice(*touchpadDevice)
		} else {
			DisableInputDevice(*touchpadDevice)
		}
	} else {
		log.Fatal("no touchpad device found")
	}
}

func FindOpenWindows() []string {
	result, err := ExecCommand("wmctrl", "-l")
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

// ExecCommand executes a shell command with the given arguments
// and returns its stdout as a []byte.
// If an error occurs the content of stderr is printed
// and an error is returned.
func ExecCommand(command string, args ...string) (string, error) {
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

// Like ExecCommand but with the possibility to add environment variables
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
