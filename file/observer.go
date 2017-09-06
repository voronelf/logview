package file

import (
	"bufio"
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/voronelf/logview/core"
	"os"
)

func NewObserver() *implObserver {
	return &implObserver{}
}

type implObserver struct {
}

var _ core.Observer = (*implObserver)(nil)

func (o *implObserver) Subscribe(ctx context.Context, filePath string, filter core.Filter) (core.Subscription, error) {
	outputCh := make(chan core.Row, 16)
	s := core.Subscription{Channel: outputCh}

	file, err := os.Open(filePath)
	if err != nil {
		return s, err
	}
	_, err = file.Seek(0, 2)
	if err != nil {
		return s, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return s, err
	}
	err = watcher.Add(filePath)
	if err != nil {
		watcher.Close()
		return s, err
	}

	go func() {
		defer watcher.Close()
		reader := bufio.NewReaderSize(file, 16384)
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					readToEOF(reader, filter, outputCh)
				}

			case err := <-watcher.Errors:
				o.processErr(err, outputCh)

			case <-ctx.Done():
				return
			}
		}
	}()

	return s, nil
}

func (o *implObserver) processErr(err error, outputCh chan<- core.Row) {
	row := core.Row{
		Data: map[string]interface{}{},
		Err:  err,
	}
	outputCh <- row
}
