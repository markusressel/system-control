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
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func isMuted(channel string) bool {
	result, err := execCommand("amixer", "get", channel)
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("\\[(on|off)\\]")
	match := r.FindString(string(result))
	return match == "[off]"
}

func setMuted(channel string, muted bool) {
	var state string
	if muted {
		state = "off"
	} else {
		state = "on"
	}

	_, err := execCommand("amixer", "set", channel, state)
	if err != nil {
		log.Fatal(err)
	}
}

func getVolume(channel string) int {
	result, err := execCommand("amixer", "get", channel)
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile("\\[\\d+%]")
	match := r.FindString(string(result))
	match = match[1 : len(match)-2]
	volume, err := strconv.Atoi(match)
	if err != nil {
		log.Fatal(err)
	}
	return volume
}

func setVolume(channel string, volume int) {
	_, err := execCommand("amixer", "set", channel, strconv.Itoa(volume)+"%")
	if err != nil {
		log.Fatal(err)
	}
}

func findSink(text string) int {
	result, err := execCommand("pacmd", "list-sinks")
	if err != nil {
		log.Fatal(err)
	}

	// find the wanted text
	i := strings.Index(result, text)
	if i == -1 {
		log.Fatalf("Substring %s not found", text)
	}

	substring := result[0 : i+len(text)]
	ri := regexp.MustCompile("index: \\d+")
	matches := ri.FindAllString(substring, -1)
	match := matches[len(matches)-1]

	rd := regexp.MustCompile("\\d+")
	sinkIndex := rd.FindString(match)

	index, err := strconv.Atoi(sinkIndex)
	if err != nil {
		log.Fatal(err)
	}

	return index
}

// Switches the default sink and moves all existing sink inputs to the target sink
func switchSink(index int) {
	indexString := strconv.Itoa(index)
	_, err := execCommand("pacmd", "set-default-sink", indexString)
	if err != nil {
		log.Fatal(err)
	}

	result, err := execCommand("pacmd", "list-sink-inputs", indexString)
	if err != nil {
		log.Fatal(err)
	}

	ri := regexp.MustCompile("index: (\\d+)")
	matches := ri.FindAllStringSubmatch(result, -1)

	for i := range matches {
		inputIdx := matches[i][1]
		_, err := execCommand("pacmd", "move-sink-input", inputIdx, indexString)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getBrightness() int {
	env := []string{"DISPLAY:=0"}
	result, err := execCommandEnv(env, true, "xbacklight")
	if err != nil {
		log.Fatal(err)
	}
	b, err := strconv.Atoi(result)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// Adjusts the brightness of the display
func adjustBrightness(change int) {
	env := []string{"DISPLAY:=0"}
	var command string
	if change >= 0 {
		command = "-inc"
	} else {
		command = "-dec"
		change = -change
	}
	_, err := execCommandEnv(env, true, "xbacklight", command, strconv.Itoa(change), "-steps", "1", "-time", "0")
	if err != nil {
		log.Fatal(err)
	}

}

// Executes a shell command with the given arguments
// and returns its stdout as a []byte.
// If an error occurs the content of stderr is printed
// and an error is returned.
func execCommand(command string, args ...string) (string, error) {
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

	return string(stdout.Bytes()), nil
}

// Like execCommand but with the possibility to add environment variables
// to the executed process.
func execCommandEnv(env []string, attach bool, command string, args ...string) (string, error) {
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
