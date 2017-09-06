package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/voronelf/logview/core"
	"io"
)

func readToEOF(reader *bufio.Reader, filter core.Filter, outputCh chan<- core.Row) {
	for {
		row := core.Row{
			Data: make(map[string]interface{}, 8),
		}
		line, isPrefix, readErr := reader.ReadLine()
		if isPrefix {
			for isPrefix {
				_, isPrefix, _ = reader.ReadLine()
			}
			row.Err = errors.New("Line is overlong. Reading was stopped")
			outputCh <- row
			return
		}
		if readErr != nil && readErr != io.EOF {
			row.Err = errors.New(readErr.Error() + ". Reading was stopped")
			outputCh <- row
			return
		}
		if len(line) > 0 {
			err := json.Unmarshal(line, &row.Data)
			if err == nil {
				if filter.Match(row) {
					outputCh <- row
				}
			} else {
				row.Err = err
				outputCh <- row
			}
		}
		if readErr == io.EOF {
			break
		}
	}
}
