package command

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/voronelf/logview/core"
	"sync"
	"testing"
	"time"
)

func newWatchForTest() (*Watch, chan<- struct{}) {
	shutdownCh := make(chan struct{})
	return &Watch{
		RowProvider:   &core.MockRowProvider{},
		FilterFactory: &core.MockFilterFactory{},
		Formatter:     &core.MockFormatter{},
		Ui:            &cli.MockUi{},
		ShutdownCh:    shutdownCh,
	}, shutdownCh
}

func TestWatch_Run_File(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	rowsChan := make(chan core.Row, 2)
	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}
	mockFilter := &core.MockFilter{}
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, "someFile").Return((<-chan core.Row)(rowsChan), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row).Return("SomeData").Twice()

	go cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})
	time.Sleep(time.Millisecond)
	rowsChan <- row
	time.Sleep(time.Millisecond)
	rowsChan <- row
	time.Sleep(time.Millisecond)

	mockProvider.AssertExpectations(t)
	mockFilterFactory.AssertExpectations(t)
	mockFilter.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	expectedOutput := messageWatchFile("someFile", "someFilter") + "\nSomeData\nSomeData\n"
	assert.Equal(t, expectedOutput, cmd.Ui.(*cli.MockUi).OutputWriter.String())
}

func TestWatch_Run_Stdin(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	cmd.Stdin = &bytes.Buffer{}
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	rowsCh := make(chan core.Row, 2)
	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}
	mockFilter := &core.MockFilter{}
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchOpenedStream", mock.Anything, cmd.Stdin).Return((<-chan core.Row)(rowsCh), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row).Return("SomeData").Twice()

	go cmd.Run([]string{"-c", "someFilter"})
	time.Sleep(time.Millisecond)
	rowsCh <- row
	time.Sleep(time.Millisecond)
	rowsCh <- row
	time.Sleep(time.Millisecond)

	mockProvider.AssertExpectations(t)
	mockFilterFactory.AssertExpectations(t)
	mockFilter.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	expectedOutput := messageWatchStdin("someFilter") + "\nSomeData\nSomeData\n"
	assert.Equal(t, expectedOutput, cmd.Ui.(*cli.MockUi).OutputWriter.String())
}

func TestWatch_Run_Shutdown(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()

	cmd.FilterFactory.(*core.MockFilterFactory).On("NewFilter", mock.Anything).Return(&core.MockFilter{}, nil)
	cmd.RowProvider.(*core.MockRowProvider).On("WatchFileChanges", mock.Anything, mock.Anything, mock.Anything).Return(make(<-chan core.Row), nil)

	done := false
	cond := sync.NewCond(&sync.Mutex{})
	go func() {
		cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})
		cond.L.Lock()
		done = true
		cond.L.Unlock()
		cond.Broadcast()
	}()

	close(shutdownCh)
	cond.L.Lock()
	cond.Wait()
	assert.True(t, done)
	cond.L.Unlock()
}

func TestWatch_Run_SubscribeError(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	mockFilterFactory.On("NewFilter", "someFilter").Return(&core.MockFilter{}, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, "someFile").Return(nil, errors.New("Some error")).Once()

	cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})

	mockFilterFactory.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	assert.Equal(t, "Some error\n", cmd.Ui.(*cli.MockUi).ErrorWriter.String())
}

func TestWatch_Run_FilterFactoryError(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)
	mockFilterFactory.On("NewFilter", "someFilter").Return(nil, errors.New("Some error")).Once()

	cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})

	mockFilterFactory.AssertExpectations(t)
	assert.Equal(t, "Some error\n", cmd.Ui.(*cli.MockUi).ErrorWriter.String())
}
