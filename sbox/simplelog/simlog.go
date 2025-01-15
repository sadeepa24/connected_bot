package simplelog

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type SimpleLog struct {
	//buf *bytes.Buffer
	//ctx context.Context
	path   string
	writer *bufio.Writer
}

func Newsimpllogger(ctx context.Context, path string) (*SimpleLog, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return nil, err
	}
	return &SimpleLog{
		writer: bufio.NewWriter(file),
		path:   path,
	}, nil

}

func (s *SimpleLog) Info(info ...any) {

	var writestring string

	for _, item := range info {
		writestring = fmt.Sprintf(writestring+"%v", item)
	}
	s.writer.WriteString(writestring + "\n")
	s.autosync()
}

func (s *SimpleLog) Sync() {
	s.writer.Flush()

}

func (s *SimpleLog) autosync() {
	if s.writer.Available()+1000 > s.writer.Size() {
		s.writer.Flush()
	}
}
