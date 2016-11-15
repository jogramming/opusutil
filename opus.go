package opusutil

import (
	"errors"
	"time"
)

// Header represents the opus packet's TOC plus extra information depending on config
type Header struct {
	Config    *Config
	NumFrames int
	Stereo    bool
}

// FullDuration returns the full duration of the opus packet (frameduration * number of frames)
func (h *Header) FullDuration() time.Duration {
	return time.Duration(h.NumFrames) * h.Config.FrameDuration
}

// DecodeHeader parses the TOC byte and more depending on config
// will return an error is opus packet is invalid to the spec.
func DecodeHeader(packet []byte) (header *Header, err error) {
	if len(packet) < 1 {
		err = errors.New("Invalid opus packet, len < 1")
		return
	}

	toc := packet[0]
	framesBits := toc & 0x3          // Framecount in 0-1
	stereo := (toc>>2)&1 != 0        // Stereo in 2
	ConfigIndex := (toc >> 3) & 0x1f // Config in 3-7

	config := ConfigTable[ConfigIndex]

	// Read number of frames depending on framesBits
	numFrames := -1
	switch framesBits {
	case 0:
		numFrames = 1
	case 1, 2:
		numFrames = 2
	case 3: // Signaled number of frames (upto max 120ms of audio)

		// This packet requires 2 bytes at min
		if len(packet) < 2 {
			err = errors.New("Invalid opus packet, len < 2 && c = 3")
			return
		}

		numFrames = int(packet[1] & 0x3f) // Count in 0-5
		if numFrames < 1 {
			err = errors.New("Invalid opus packet, framcount < 1")
		}
	}

	header = &Header{
		Config:    config,
		NumFrames: numFrames,
		Stereo:    stereo,
	}
	return
}

type Codec int

const (
	SILK = iota
	CELT
	Hybrid
)

// Config represents the config bits in the TOC byte
type Config struct {
	Codec         Codec
	FrameDuration time.Duration
	Bandwidth     *Bandwidth
}

type Bandwidth struct {
	Bandwidth  int
	SampleRate int
}

var (
	NB  = &Bandwidth{4, 8}   // Narrow band
	MB  = &Bandwidth{6, 12}  // Medium
	WB  = &Bandwidth{8, 16}  // Wide
	SWB = &Bandwidth{12, 24} // Super-wide
	FB  = &Bandwidth{20, 48} // Full
)

// Opus config mapping table
var ConfigTable = [32]*Config{
	// Silk
	{SILK, 10000 * time.Microsecond, NB}, {SILK, 20000 * time.Microsecond, NB}, {SILK, 40000 * time.Microsecond, NB}, {SILK, 60000 * time.Microsecond, NB},
	{SILK, 10000 * time.Microsecond, MB}, {SILK, 20000 * time.Microsecond, MB}, {SILK, 40000 * time.Microsecond, MB}, {SILK, 60000 * time.Microsecond, MB},
	{SILK, 10000 * time.Microsecond, WB}, {SILK, 20000 * time.Microsecond, WB}, {SILK, 40000 * time.Microsecond, WB}, {SILK, 60000 * time.Microsecond, WB},

	// Hybrid
	{Hybrid, 10000 * time.Microsecond, SWB}, {Hybrid, 20000 * time.Microsecond, SWB},
	{Hybrid, 10000 * time.Microsecond, FB}, {Hybrid, 20000 * time.Microsecond, FB},

	// CELT
	{CELT, 2500 * time.Microsecond, NB}, {CELT, 5000 * time.Microsecond, NB}, {CELT, 10000 * time.Microsecond, NB}, {CELT, 20000 * time.Microsecond, NB},
	{CELT, 2500 * time.Microsecond, WB}, {CELT, 5000 * time.Microsecond, WB}, {CELT, 10000 * time.Microsecond, WB}, {CELT, 20000 * time.Microsecond, WB},
	{CELT, 2500 * time.Microsecond, SWB}, {CELT, 5000 * time.Microsecond, SWB}, {CELT, 10000 * time.Microsecond, SWB}, {CELT, 20000 * time.Microsecond, SWB},
	{CELT, 2500 * time.Microsecond, FB}, {CELT, 5000 * time.Microsecond, FB}, {CELT, 10000 * time.Microsecond, FB}, {CELT, 20000 * time.Microsecond, FB},
}
