package provider

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

func createFileWithJson(t *testing.T) (*os.File, func()) {
	fd, err := ioutil.TempFile("", "go_test_")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.Write([]byte("{\"field\": \"1\"}\n{\"field\": \"2\"}\n{\"field\": \"3\"}\n"))
	if err != nil {
		t.Fatal(err)
	}
	return fd, func() { fd.Close(); os.Remove(fd.Name()) }
}

func startWatchFileChanges(t *testing.T) (io.Writer, *sync.Map, func()) {
	tempFile, delTempFile := createFileWithJson(t)
	ctx, cancelCtx := context.WithCancel(context.Background())

	rowsChan, err := NewRowProvider().WatchFileChanges(ctx, tempFile.Name())
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	results := &sync.Map{}
	go func() {
		i := 0
		for {
			select {
			case row, ok := <-rowsChan:
				if !ok {
					return
				}
				results.Store(i, row)
				i++
			case <-ctx.Done():
				return
			}
		}
	}()
	time.Sleep(time.Millisecond)
	return tempFile, results, func() { delTempFile(); cancelCtx() }
}

func TestRowProvider_WatchFileChanges_oneRow(t *testing.T) {
	tempFile, results, stopWatch := startWatchFileChanges(t)
	defer stopWatch()

	tempFile.Write([]byte("{\"field\": \"456\"}\n"))
	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "456", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(1)
	assert.False(t, loaded)
}

func TestRowProvider_WatchFileChanges_twoRows(t *testing.T) {
	tempFile, results, stopWatch := startWatchFileChanges(t)
	defer stopWatch()

	tempFile.Write([]byte("{\"field\": 777}\n{\"field\": 888.99}\n"))
	time.Sleep(2 * time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, float64(777), value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	value, loaded = results.Load(1)
	if assert.True(t, loaded) {
		assert.Equal(t, float64(888.99), value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(2)
	assert.False(t, loaded)
}

func TestRowProvider_WatchFileChanges_partialWrite(t *testing.T) {
	tempFile, results, stopWatch := startWatchFileChanges(t)
	defer stopWatch()

	tempFile.Write([]byte("{\"field\": \"45"))
	time.Sleep(time.Millisecond)
	_, loaded := results.Load(0)
	assert.False(t, loaded)

	tempFile.Write([]byte("6\"}"))
	time.Sleep(time.Millisecond)
	_, loaded = results.Load(0)
	assert.False(t, loaded)

	tempFile.Write([]byte("\n"))
	time.Sleep(2 * time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "456", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(1)
	assert.False(t, loaded)
}

func TestRowProvider_WatchFileChanges_emptyLine(t *testing.T) {
	tempFile, results, stopWatch := startWatchFileChanges(t)
	defer stopWatch()

	tempFile.Write([]byte("\n"))
	time.Sleep(time.Millisecond)
	_, loaded := results.Load(0)
	assert.False(t, loaded)
}

func TestRowProvider_WatchFileChanges_Err(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	_, err := NewRowProvider().WatchFileChanges(ctx, "notExistsFilePath")
	assert.NotNil(t, err)
}

func startReadFileTail(t *testing.T, countBytes int, bytesToAddInFile []byte) (*sync.Map, func()) {
	tempFile, delTempFile := createFileWithJson(t)
	ctx, cancelCtx := context.WithCancel(context.Background())

	if len(bytesToAddInFile) > 0 {
		tempFile.Write(bytesToAddInFile)
	}

	rowsChan, err := NewRowProvider().ReadFileTail(ctx, tempFile.Name(), int64(countBytes))
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	results := &sync.Map{}
	go func() {
		i := 0
		for {
			select {
			case row, ok := <-rowsChan:
				if !ok {
					return
				}
				results.Store(i, row)
				i++
			case <-ctx.Done():
				return
			}
		}
	}()
	time.Sleep(time.Millisecond)
	return results, func() { delTempFile(); cancelCtx() }
}

func TestRowProvider_ReadFileTail(t *testing.T) {
	results, stopWatch := startReadFileTail(t, 37, []byte{})
	defer stopWatch()

	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "2", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	value, loaded = results.Load(1)
	if assert.True(t, loaded) {
		assert.Equal(t, "3", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(2)
	assert.False(t, loaded)
}

func TestRowProvider_ReadFileTail_allFile(t *testing.T) {
	results, stopWatch := startReadFileTail(t, 0, []byte{})
	defer stopWatch()

	time.Sleep(time.Millisecond)
	_, loaded := results.Load(0)
	assert.True(t, loaded)
	_, loaded = results.Load(1)
	assert.True(t, loaded)
	_, loaded = results.Load(2)
	assert.True(t, loaded)
	_, loaded = results.Load(3)
	assert.False(t, loaded)
}

func TestRowProvider_ReadFileTail_withoutDivider(t *testing.T) {
	lastLine := []byte("{\"field\": \"456\"}")
	results, stopWatch := startReadFileTail(t, len(lastLine)+2, lastLine)
	defer stopWatch()

	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "456", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(1)
	assert.False(t, loaded)
}

func TestRowProvider_ReadFileTail_Err(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	_, err := NewRowProvider().ReadFileTail(ctx, "notExistsFilePath", 123)
	assert.NotNil(t, err)
}

func startWatchOpenedStream(t *testing.T) (io.Writer, *sync.Map, func()) {
	pipeReader, pipeWriter := io.Pipe()
	ctx, cancelCtx := context.WithCancel(context.Background())

	rowsChan, err := NewRowProvider().WatchOpenedStream(ctx, pipeReader)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	results := &sync.Map{}
	go func() {
		i := 0
		for {
			select {
			case row, ok := <-rowsChan:
				if !ok {
					return
				}
				results.Store(i, row)
				i++
			case <-ctx.Done():
				return
			}
		}
	}()
	time.Sleep(time.Millisecond)
	return pipeWriter, results, func() { pipeReader.Close(); cancelCtx() }
}

func TestRowProvider_WatchOpenedStream_oneRow(t *testing.T) {
	w, results, stopWatch := startWatchOpenedStream(t)
	defer stopWatch()

	w.Write([]byte("{\"field\": \"456\"}\n"))
	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "456", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(1)
	assert.False(t, loaded)
}

func TestRowProvider_WatchOpenedStream_twoRows(t *testing.T) {
	w, results, stopWatch := startWatchOpenedStream(t)
	defer stopWatch()

	w.Write([]byte("{\"field\": 777}\n{\"field\": 888.99}\n"))
	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, float64(777), value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	value, loaded = results.Load(1)
	if assert.True(t, loaded) {
		assert.Equal(t, float64(888.99), value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(2)
	assert.False(t, loaded)
}

func TestRowProvider_WatchOpenedStream_partialWrite(t *testing.T) {
	tempFile, results, stopWatch := startWatchOpenedStream(t)
	defer stopWatch()

	tempFile.Write([]byte("{\"field\": \"45"))
	time.Sleep(time.Millisecond)
	_, loaded := results.Load(0)
	assert.False(t, loaded)

	tempFile.Write([]byte("6\"}"))
	time.Sleep(time.Millisecond)
	_, loaded = results.Load(0)
	assert.False(t, loaded)

	tempFile.Write([]byte("\n"))
	time.Sleep(time.Millisecond)
	value, loaded := results.Load(0)
	if assert.True(t, loaded) {
		assert.Equal(t, "456", value.(core.Row).Data["field"])
		assert.Nil(t, value.(core.Row).Err)
	}
	_, loaded = results.Load(1)
	assert.False(t, loaded)
}

func TestRowProvider_WatchOpenedStream_emptyLine(t *testing.T) {
	tempFile, results, stopWatch := startWatchOpenedStream(t)
	defer stopWatch()

	tempFile.Write([]byte("\n"))
	time.Sleep(time.Millisecond)
	_, loaded := results.Load(0)
	assert.False(t, loaded)
}
