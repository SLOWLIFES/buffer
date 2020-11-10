package buffer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"time"
)

type FrameType string

const (
	FrameTime      FrameType = "time"
	FrameDelimiter           = "delim"
	FrameLen                 = "len"
)

type Buffer struct {
	ReadWriter     io.ReadWriteCloser `json:"read_writer"`
	FrameType      FrameType          `json:"frame_type"`
	FrameTime      time.Duration      `json:"frame_time"`
	FrameDelimiter string             `json:"frame_delimiter"`
	FrameLen       int                `json:"frame_len"`
	IsRun          bool               `json:"is_run"`
	Error          error              `json:"error"`
	timeBuffer     *timeBuffer
	delimScanner   *bufio.Scanner
	lenBuffer      *bytes.Buffer
}

func (buff *Buffer) ReadFrame() ([]byte, error) {
	switch buff.FrameType {
	case FrameTime:
		return buff.readFrameWithTime()
	case FrameDelimiter:
		return buff.readFrameWithDelim()
	case FrameLen:
		return buff.readFrameWithLen()
	default:
		return nil, errors.New("frame type is not defined")
	}
}

func (buff *Buffer) Write(p []byte) (n int, err error) {
	return buff.ReadWriter.Write(p)
}

func (buff *Buffer) Close() {
	buff.IsRun = false
	_ = buff.ReadWriter.Close()
}
