package file

import (
	"bufio"
	"github.com/voronelf/logview/core"
	"os"
)

func NewFileReader() *reader {
	return &reader{}
}

type reader struct {
}

var _ core.FileReader = (*reader)(nil)

func (*reader) ReadTail(filePath string, b int64, filter core.Filter) (<-chan core.Row, error) {
	filteredRowsCh := make(chan core.Row, 1)
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	if b > 0 {
		fd.Seek(-b, 2)
	}
	go func() {
		defer fd.Close()
		reader := bufio.NewReaderSize(fd, maxBytesInRow)
		if b > 0 {
			reader.ReadLine()
		}
		readToEOF(reader, filter, filteredRowsCh)
		close(filteredRowsCh)
	}()
	return filteredRowsCh, nil
}
