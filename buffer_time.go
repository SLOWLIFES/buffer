package buffer

import (
	"bytes"
	"container/list"
	"io"
	"sync"
	"time"
)

type timeBuffer struct {
	Queue    *list.List    `json:"queue"`
	Buffer   *bytes.Buffer `json:"buffer"`
	LastTime time.Time     `json:"last_time"`
	lock     sync.Mutex
}

func (tb *timeBuffer) Push(v []byte) {
	defer tb.lock.Unlock()
	tb.lock.Lock()
	tb.Queue.PushBack(v)
}

func (tb *timeBuffer) Pop() []byte {
	defer tb.lock.Unlock()
	tb.lock.Lock()
	value := tb.Queue.Front()
	tb.Queue.Remove(value)
	return value.Value.([]byte)
}

func NewBufferWithTime(rwc io.ReadWriteCloser, td time.Duration) *Buffer {
	buff := Buffer{
		ReadWriter: rwc,
		FrameType:  FrameTime,
		FrameTime:  td,
		timeBuffer: &timeBuffer{
			Queue:    list.New(),
			Buffer:   bytes.NewBuffer([]byte{}),
			LastTime: time.Now(),
		},
		IsRun: true,
	}

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
			//log.Println("debuf:",string(p[0:n]))
			buff.timeBuffer.Buffer.Write(p[0:n])
			buff.timeBuffer.LastTime = time.Now()
		}
	}()

	go func() {
		defer func() {
			recover()
		}()

		for buff.IsRun {
			data := buff.timeBuffer.Buffer.Bytes()
			if time.Now().After(buff.timeBuffer.LastTime.Add(buff.FrameTime)) && len(data) > 0 {
				buff.timeBuffer.Push(data)
				buff.timeBuffer.Buffer.Reset()
			}
			time.Sleep(time.Millisecond / 2)
		}
	}()
	return &buff
}

func (buff *Buffer) readFrameWithTime() ([]byte, error) {
	for {
		if buff.timeBuffer.Queue.Len() > 0 {
			value := buff.timeBuffer.Pop()
			return value, nil
		}
		if !buff.IsRun {
			return nil, buff.Error
		}
		time.Sleep(time.Millisecond)
	}
}
