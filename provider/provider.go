package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/voronelf/logview/core"
	"io"
	"os"
)

func NewRowProvider() *rowProvider {
	return &rowProvider{}
}

type rowProvider struct {
}

var _ core.RowProvider = (*rowProvider)(nil)

func (r *rowProvider) WatchFileChanges(ctx context.Context, filePath string) (<-chan core.Row, error) {
	outputCh := make(chan core.Row, 16)
	file, err := os.Open(filePath)
	if err != nil {
		return outputCh, err
	}
	_, err = file.Seek(0, 2)
	if err != nil {
		return outputCh, err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return outputCh, err
	}
	err = watcher.Add(filePath)
	if err != nil {
		watcher.Close()
		return outputCh, err
	}
	go func() {
		defer func() {
			watcher.Close()
			file.Close()
		}()
		reader := newReaderIgnoreEOF(file, outputCh)
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					reader.readFragment(ctx)
				}

			case err := <-watcher.Errors:
				r.sendRowWithErr(err, outputCh)

			case <-ctx.Done():
				return
			}
		}
	}()
	return outputCh, nil
}

func (r *rowProvider) WatchOpenedStream(ctx context.Context, stream io.Reader) (<-chan core.Row, error) {
	filteredRowsCh := make(chan core.Row, 1)
	go func() {
		reader := bufio.NewReaderSize(stream, maxBytesInRow)
		readUntilEOF(ctx, reader, filteredRowsCh)
		close(filteredRowsCh)
	}()
	return filteredRowsCh, nil
}

func (r *rowProvider) ReadFileTail(ctx context.Context, filePath string, countBytes int64) (<-chan core.Row, error) {
	outputCh := make(chan core.Row, 16)
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	if countBytes > 0 {
		fd.Seek(-countBytes, 2)
	}
	go func() {
		defer fd.Close()
		reader := bufio.NewReaderSize(fd, maxBytesInRow)
		if countBytes > 0 {
			reader.ReadLine()
		}
		readUntilEOF(ctx, reader, outputCh)
		close(outputCh)
	}()
	return outputCh, nil
}

func (r *rowProvider) sendRowWithErr(err error, outputCh chan<- core.Row) {
	row := core.Row{
		Data: map[string]interface{}{},
		Err:  err,
	}
	outputCh <- row
}

const maxBytesInRow = 16384

func readUntilEOF(ctx context.Context, reader *bufio.Reader, outputCh chan<- core.Row) {
	for ctx.Err() == nil {
		line, isPrefix, err := reader.ReadLine()
		if isPrefix {
			for isPrefix {
				_, isPrefix, err = reader.ReadLine()
				if err != nil && err != io.EOF {
					outputCh <- core.Row{Err: err}
					return
				}
			}
			outputCh <- core.Row{Err: errors.New("line is overlong")}
			continue
		}
		if err == io.EOF {
			if len(line) > 0 {
				outputCh <- createRow(line)
			}
			return
		}
		if err != nil {
			outputCh <- core.Row{Err: err}
			return
		}
		if len(line) > 0 {
			outputCh <- createRow(line)
		}
	}
}

func createRow(line []byte) core.Row {
	row := core.Row{
		Data: make(map[string]interface{}, 8),
	}
	row.Err = json.Unmarshal(line, &row.Data)
	return row
}

func newReaderIgnoreEOF(r io.Reader, outputCh chan<- core.Row) *readerIgnoreEOF {
	return &readerIgnoreEOF{
		rd:       bufio.NewReaderSize(r, maxBytesInRow),
		outputCh: outputCh,
	}
}

type readerIgnoreEOF struct {
	buf      [maxBytesInRow]byte
	pos      int
	rd       *bufio.Reader
	outputCh chan<- core.Row
}

func (r *readerIgnoreEOF) readFragment(ctx context.Context) {
	for ctx.Err() == nil {
		slice, err := r.rd.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			r.pos = 0
			for err == bufio.ErrBufferFull {
				_, err = r.rd.ReadSlice('\n')
				if err != nil && err != io.EOF {
					r.outputCh <- core.Row{Err: err}
					return
				}
			}
			r.outputCh <- core.Row{Err: errors.New("line is overlong")}
			continue
		}
		if err == io.EOF {
			e := r.saveToBuf(slice)
			if e != nil {
				r.outputCh <- core.Row{Err: e}
			}
			return
		}
		if err != nil {
			r.outputCh <- core.Row{Err: err}
			return
		}
		if r.pos == 0 {
			if len(slice) > 2 { // slice contains '\n' or '\r\n'
				r.outputCh <- createRow(slice)
			}
		} else {
			e := r.saveToBuf(slice)
			if e != nil {
				r.outputCh <- core.Row{Err: e}
				return
			}
			r.outputCh <- createRow(r.buf[:r.pos])
			r.pos = 0
		}
	}
}

func (r *readerIgnoreEOF) saveToBuf(slice []byte) error {
	sliceLength := len(slice)
	if sliceLength == 0 {
		return nil
	}
	copied := copy(r.buf[r.pos:], slice)
	if copied != sliceLength {
		return errors.New("reader buffer is full")
	}
	r.pos += copied
	return nil
}
