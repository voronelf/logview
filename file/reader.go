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

func (*reader) ReadAll(filePath string, filter core.Filter) (<-chan core.Row, error) {
	filteredRowsCh := make(chan core.Row, 1)
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	go func() {
		defer fd.Close()
		reader := bufio.NewReaderSize(fd, 16384)
		readToEOF(reader, filter, filteredRowsCh)
		close(filteredRowsCh)
	}()
	return filteredRowsCh, nil
}
