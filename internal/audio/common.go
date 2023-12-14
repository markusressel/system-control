package audio

func IsHeadphoneConnected() bool {
	// TODO:
	return false
}

// CalculateAppropriateVolumeChange calculates an appropriate amount of volume
// change when the user did not specify a specific value.
// Expects "current" to be a value between 0 and 100.
func CalculateAppropriateVolumeChange(current float64, increase bool) float64 {
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
