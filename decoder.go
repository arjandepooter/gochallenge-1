package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// Byte offsets:
// 0, 6: File header string: SPLICE
// 6, 8: Content size int64
// 14, 32: Version string
// 46, 4: Tempo float
// 50, size - 36: []Tracks
// Track byte offsets:
// 0, 4: ID int32
// 4, 1: length of track name int8
// 5, length: track name string
// 5 + length, 16: steps 00 or 01
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header, err := readHeader(f)
	if header != "SPLICE" {
		return nil, fmt.Errorf("Invalid file header, expected SPLICE, got %s", header)
	}

	size, err := readContentSize(f)
	if err != nil {
		return nil, err
	}

	version, err := readVersion(f)
	if err != nil {
		return nil, err
	}
	p.Version = version
	size -= 32

	tempo, err := readTempo(f)
	if err != nil {
		return nil, err
	}
	p.Tempo = tempo
	size -= 4

	var tracks []*Track
	for size > 0 {
		track, err := readTrack(f, &size)

		if err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}
	p.Tracks = tracks

	return p, nil
}

func readHeader(file io.Reader) (string, error) {
	buf := make([]byte, 6)
	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func readContentSize(file io.Reader) (int64, error) {
	var size int64
	err := binary.Read(file, binary.BigEndian, &size)

	if err != nil {
		return 0, err
	}

	return size, nil
}

func readVersion(file io.Reader) (string, error) {
	buf := make([]byte, 32)
	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	return string(bytes.Trim(buf, "\x00")), nil
}

func readTempo(file io.Reader) (float32, error) {
	var tempo float32
	err := binary.Read(file, binary.LittleEndian, &tempo)

	if err != nil {
		return 0, err
	}

	return tempo, nil
}

func readTrack(file io.Reader, size *int64) (*Track, error) {
	track := new(Track)

	var id int32
	err := binary.Read(file, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}
	track.ID = int(id)
	*size -= 4

	var nameLength int8
	err = binary.Read(file, binary.LittleEndian, &nameLength)
	if err != nil {
		return nil, err
	}
	*size--

	buf := make([]byte, nameLength)
	file.Read(buf)
	track.Name = string(buf)
	*size -= int64(nameLength)

	var steps [16]bool
	for i := 0; i < 16; i++ {
		var buf int8
		err = binary.Read(file, binary.LittleEndian, &buf)
		if err != nil {
			return nil, err
		}

		steps[i] = (buf > 0)
	}
	*size -= 16
	track.Steps = steps

	return track, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is a representation of a single track in a pattern
type Track struct {
	ID    int
	Name  string
	Steps [16]bool
}
