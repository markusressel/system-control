package session

import (
	"testing"

	"github.com/markusressel/system-control/internal/configuration"
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

func TestLockCommandLineDefault(t *testing.T) {
	name, args := lockCommandLine(configuration.SessionLockConfig{})

	assert.Equal(t, "i3lock", name)
	assert.Equal(t, []string{"--nofork", "--show-failed-attempts", "-c130003"}, args)
}

func TestLockCommandLineWithEnvironment(t *testing.T) {
	name, args := lockCommandLine(configuration.SessionLockConfig{
		Command:     []string{"xsecurelock"},
		Environment: []string{"XSECURELOCK_PAM_SERVICE=xsecurelock"},
	})

	assert.Equal(t, "env", name)
	assert.Equal(t, []string{"XSECURELOCK_PAM_SERVICE=xsecurelock", "xsecurelock"}, args)
}

func TestLockProcessName(t *testing.T) {
	tests := []struct {
		name   string
		config configuration.SessionLockConfig
		want   string
	}{
		{"default", configuration.SessionLockConfig{}, "i3lock"},
		{"base name of the command", configuration.SessionLockConfig{Command: []string{"/usr/bin/xsecurelock"}}, "xsecurelock"},
		{"explicit", configuration.SessionLockConfig{Command: []string{"wrapper"}, ProcessName: "locker"}, "locker"},
		{"cut to the length of comm", configuration.SessionLockConfig{Command: []string{"a-very-long-locker-name"}}, "a-very-long-loc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, lockProcessName(tt.config))
		})
	}
}
