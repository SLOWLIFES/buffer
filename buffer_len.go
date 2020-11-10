package buffer

import (
	"bytes"
	"io"
	"time"
)

func NewBufferWithLen(rwc io.ReadWriteCloser, len int) *Buffer {
	buff := Buffer{
		ReadWriter: rwc,
		FrameType:  FrameLen,
		FrameLen:   len,
		IsRun:      true,
	}
	buff.lenBuffer = bytes.NewBuffer([]byte{})

	go func() {
		defer func() {
			recover()
			buff.Close()
		}()
		p := make([]byte, 256)
		for {
			n, err := buff.ReadWriter.Read(p)
			if err != nil {
				if err == io.EOF {
					time.Sleep(time.Millisecond / 2)
					continue
				}
				buff.Error = err
				break
			}
			buff.lenBuffer.Write(p[0:n])
			time.Sleep(time.Millisecond / 2)
		}

	}()

	return &buff
}

func (buff *Buffer) readFrameWithLen() ([]byte, error) {
	var byts []byte

	for len(byts) < buff.FrameLen {
		if buff.lenBuffer.Len() >= buff.FrameLen {
			byts = append(byts, buff.lenBuffer.Next(buff.FrameLen-len(byts))...)
		}
		if !buff.IsRun {
			return nil, buff.Error
		}
		time.Sleep(time.Millisecond)
	}
	return byts, nil
}
