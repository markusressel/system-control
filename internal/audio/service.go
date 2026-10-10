package audio

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
)

// optimisticWindow is how long a value we wrote ourselves is trusted over the PipeWire graph.
// The graph lags behind our own writes, so rapid scroll events must build on the value we just
// set, and a query right after a change must not report the previous state.
const optimisticWindow = 300 * time.Millisecond

type written[T any] struct {
	value T
	at    time.Time
}

// Backend is everything the Service needs from the audio system.
type Backend struct {
	Graph      func() pipewire.GraphState
	SetVolume  func(id int, volume float64) error
	SetMute    func(id int, mute bool) error
	ToggleMute func(id int) error
}

func pipewireBackend() Backend {
	return Backend{
		Graph:      pipewire.PwDump,
		SetVolume:  pipewire.WpCtlSetVolume,
		SetMute:    pipewire.WpCtlSetMute,
		ToggleMute: pipewire.WpCtlToggleMute,
	}
}

// Service implements the audio operations. All mutations are serialized by the Service,
// which also owns the (short lived) optimistic volume state.
type Service struct {
	be Backend

	// mu serializes mutations and guards the records of our own recent writes, by node id
	mu         sync.Mutex
	lastVolume map[int]written[float64]
	lastMuted  map[int]written[bool]
}

func NewService(be Backend) *Service {
	return &Service{
		be:         be,
		lastVolume: map[int]written[float64]{},
		lastMuted:  map[int]written[bool]{},
	}
}

// volumeNodeId returns the id of the node GetVolumeByName(device) reads, 0 if there is none.
func volumeNodeId(state pipewire.GraphState, device string) int {
	if device == "" {
		if node, err := state.GetDefaultSinkNode(); err == nil {
			return node.Id
		}
		return 0
	}
	if nodes := state.FindNodesByName(device); len(nodes) == 1 {
		return nodes[0].Id
	}
	return 0
}

// currentVolume returns the volume of a device, preferring our own recent write over the graph.
// Must be called with s.mu held.
func (s *Service) currentVolume(state pipewire.GraphState, device string) (float64, error) {
	volume, err := state.GetVolumeByName(device)
	if err != nil {
		return 0, err
	}
	volume = util.RoundToTwoDecimals(volume)

	if id := volumeNodeId(state, device); id != 0 {
		if w, ok := s.lastVolume[id]; ok && time.Since(w.at) < optimisticWindow {
			return w.value, nil
		}
	}
	return volume, nil
}

// ResolveTargets returns the nodes a command applies to: the given stream, the given device
// or the default sink.
func ResolveTargets(state pipewire.GraphState, device string, stream string) ([]pipewire.InterfaceNode, error) {
	var targets []pipewire.InterfaceNode
	if stream != "" {
		targets = state.FindStreamNodes(stream)
	} else if device != "" {
		targets = state.FindNodesByName(device)
	} else {
		target, err := state.GetDefaultSinkNode()
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func volumeLine(volume float64) string {
	return fmt.Sprintf("%d", int(math.Round(volume*100)))
}

func muteLine(muted bool) string {
	if muted {
		return "yes"
	}
	return "no"
}

// Volume returns the volume of the given device (default sink if empty) in percent.
func (s *Service) Volume(device string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	volume, err := s.currentVolume(s.be.Graph(), device)
	if err != nil {
		return "", err
	}
	return volumeLine(volume), nil
}

// Muted returns "yes" or "no" for the first target of the given device/stream.
func (s *Service) Muted(device string, stream string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.be.Graph()
	targets, err := ResolveTargets(state, device, stream)
	if err != nil {
		return "", err
	}
	if len(targets) == 0 {
		return "", errors.New("no matching audio target found")
	}

	muted, err := s.currentMuted(state, targets[0].Id)
	if err != nil {
		return "", err
	}
	return muteLine(muted), nil
}

// currentMuted returns the mute state of a node, preferring our own recent write over the graph.
// Must be called with s.mu held.
func (s *Service) currentMuted(state pipewire.GraphState, id int) (bool, error) {
	if w, ok := s.lastMuted[id]; ok && time.Since(w.at) < optimisticWindow {
		return w.value, nil
	}
	return state.IsMuted(id)
}

// Step increases or decreases the volume by an appropriate amount.
func (s *Service) Step(device string, stream string, increase bool) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.be.Graph()

	volume, err := s.currentVolume(state, device)
	if err != nil {
		return nil, err
	}

	targets, err := ResolveTargets(state, device, stream)
	if err != nil {
		return nil, err
	}

	change := CalculateAppropriateVolumeChange(volume*100, increase) / 100.0
	target := volume + change
	if !increase {
		target = volume - change
	}
	if target < 0.0 {
		target = 0.0
	}

	return s.apply(targets, target)
}

// Set sets the volume to a percentage.
func (s *Service) Set(device string, stream string, percent int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targets, err := ResolveTargets(s.be.Graph(), device, stream)
	if err != nil {
		return nil, err
	}

	return s.apply(targets, math.Max(float64(percent)/100.0, 0))
}

// apply must be called with s.mu held.
func (s *Service) apply(targets []pipewire.InterfaceNode, volume float64) ([]string, error) {
	rounded := util.RoundToTwoDecimals(volume)

	var lines []string
	for _, target := range targets {
		if err := s.be.SetVolume(target.Id, volume); err != nil {
			return nil, err
		}
		s.lastVolume[target.Id] = written[float64]{rounded, time.Now()}
		lines = append(lines, volumeLine(rounded))
	}
	return lines, nil
}

type MuteAction int

const (
	MuteOn MuteAction = iota
	MuteOff
	MuteToggle
)

// Mute changes the mute state of all targets and returns the resulting state per target.
func (s *Service) Mute(device string, stream string, action MuteAction) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.be.Graph()
	targets, err := ResolveTargets(state, device, stream)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, target := range targets {
		var muted bool
		switch action {
		case MuteOn:
			muted = true
			err = s.be.SetMute(target.Id, true)
		case MuteOff:
			muted = false
			err = s.be.SetMute(target.Id, false)
		case MuteToggle:
			initial, _ := s.currentMuted(state, target.Id)
			muted = !initial
			err = s.be.ToggleMute(target.Id)
		}
		if err != nil {
			return nil, err
		}
		s.lastMuted[target.Id] = written[bool]{muted, time.Now()}
		lines = append(lines, muteLine(muted))
	}
	return lines, nil
}

type savedAudioState struct {
	Device string
	Volume float64
	Muted  bool
}

func savedStateKey(deviceName string) string {
	return "audio_state_" + deviceName
}

// Save persists volume and mute state of all targets.
func (s *Service) Save(device string, stream string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.be.Graph()
	targets, err := ResolveTargets(state, device, stream)
	if err != nil {
		return err
	}

	for _, target := range targets {
		name, _ := target.GetName()
		volume, _ := state.GetVolumeByName(name)
		muted, _ := state.IsMuted(target.Id)

		data := savedAudioState{Device: name, Volume: volume, Muted: muted}
		if err := persistence.SaveStruct(savedStateKey(name), &data); err != nil {
			return err
		}
	}
	return nil
}

// Restore applies a previously saved state, returning mute and volume lines per restored target.
func (s *Service) Restore(device string, stream string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.be.Graph()
	targets, err := ResolveTargets(state, device, stream)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, target := range targets {
		name, _ := target.GetName()
		data := savedAudioState{}
		if err := persistence.ReadStruct(savedStateKey(name), &data); err != nil {
			continue // nothing saved for this target
		}

		if err := s.be.SetMute(target.Id, data.Muted); err != nil {
			return nil, err
		}
		if err := s.be.SetVolume(target.Id, data.Volume); err != nil {
			return nil, err
		}
		rounded := util.RoundToTwoDecimals(data.Volume)
		now := time.Now()
		s.lastMuted[target.Id] = written[bool]{data.Muted, now}
		s.lastVolume[target.Id] = written[float64]{rounded, now}
		lines = append(lines, muteLine(data.Muted), volumeLine(rounded))
	}
	return lines, nil
}

// SwitchSink makes the single sink matching the search string the default sink.
func SwitchSink(search string) error {
	state := pipewire.PwDump()

	var sinks []pipewire.InterfaceNode
	for _, node := range state.FindNodesByName(search) {
		if mediaClass, _ := node.GetMediaClass(); mediaClass == pipewire.MediaClassAudioSink {
			sinks = append(sinks, node)
		}
	}

	if len(sinks) == 0 {
		return errors.New("no sink found")
	}
	if len(sinks) > 1 {
		names := make([]string, len(sinks))
		for i, node := range sinks {
			name, _ := node.GetName()
			description, _ := node.GetDescription()
			names[i] = fmt.Sprintf("%s (%s)", name, description)
		}
		return fmt.Errorf("ambiguous sink name, found: %v", names)
	}
	return state.SwitchSinkTo(sinks[0])
}
