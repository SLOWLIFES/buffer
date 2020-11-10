package buffer

import (
	"bufio"
	"io"
	"strings"
)

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func NewBufferWithDelimiter(rwc io.ReadWriteCloser, delim string) *Buffer {
	buff := Buffer{
		ReadWriter:     rwc,
		FrameType:      FrameDelimiter,
		FrameDelimiter: delim,
		IsRun:          true,
	}
	scanner := bufio.NewScanner(buff.ReadWriter)
	scanner.Split(buff.scanDelimiter)
	buff.delimScanner = scanner
	return &buff
}

func (buff *Buffer) scanDelimiter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := strings.Index(string(data), buff.FrameDelimiter); i >= 0 {
		// We have a full newline-terminated line.
		return i + len(buff.FrameDelimiter), dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func (buff *Buffer) readFrameWithDelim() ([]byte, error) {

	for {
		ok := buff.delimScanner.Scan()
		if ok {
			return buff.delimScanner.Bytes(), nil
		}
		buff.Error = buff.delimScanner.Err()
		if buff.delimScanner.Err() != nil {
			return nil, buff.delimScanner.Err()
		}
		if !buff.IsRun {
			return nil, io.ErrClosedPipe
		}
	}

}
