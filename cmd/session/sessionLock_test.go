package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionCommandOtherUser(t *testing.T) {
	name, args := sessionCommand("alice", "root", ":1", "/usr/bin", "xset", "dpms", "force", "off")

	assert.Equal(t, "sudo", name)
	assert.Equal(t, []string{"-u", "alice", "env", "DISPLAY=:1", "PATH=/usr/bin", "xset", "dpms", "force", "off"}, args)
}

func TestSessionCommandSameUser(t *testing.T) {
	name, args := sessionCommand("alice", "alice", ":0", "", "xset", "s", "0", "0")

	assert.Equal(t, "env", name)
	assert.Equal(t, []string{"DISPLAY=:0", "xset", "s", "0", "0"}, args)
}
