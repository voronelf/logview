package command

import (
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/voronelf/logview/core"
	"testing"
	"time"
)

func newTailForTest() (*Tail, chan<- struct{}) {
	shutdownCh := make(chan struct{})
	return &Tail{
		RowProvider:   &core.MockRowProvider{},
		FilterFactory: &core.MockFilterFactory{},
		Formatter:     &core.MockFormatter{},
		Ui:            &cli.MockUi{},
		ShutdownCh:    shutdownCh,
	}, shutdownCh
}

func TestTail_Run(t *testing.T) {
	cmd, shutdownCh := newTailForTest()
	defer close(shutdownCh)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)

	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}
	channel := make(chan core.Row, 2)
	channel <- row
	channel <- row
	close(channel)
	mockFilter := &core.MockFilter{}
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("ReadFileTail", mock.Anything, "someFile", int64(123)).Return((<-chan core.Row)(channel), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row, core.DefaultFormatParams()).Return("SomeData").Twice()

	cmd.Run([]string{"-f", "someFile", "-b", "123", "-c", "someFilter"})

	mockProvider.AssertExpectations(t)
	mockFilterFactory.AssertExpectations(t)
	mockFilter.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	assert.Equal(t, "SomeData\nSomeData\n", cmd.Ui.(*cli.MockUi).OutputWriter.String())
}

func TestTail_Run_WithDateInFilePath(t *testing.T) {
	cmd, shutdownCh := newTailForTest()
	defer close(shutdownCh)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)

	incomingFile := "someFile_@today@.log"
	expectedFile := "someFile_" + time.Now().UTC().Format("2006-01-02") + ".log"

	channel := make(chan core.Row, 2)
	close(channel)
	mockFilterFactory.On("NewFilter", "someFilter").Return(&core.MockFilter{}, nil).Once()
	mockProvider.On("ReadFileTail", mock.Anything, expectedFile, int64(123)).Return((<-chan core.Row)(channel), nil).Once()

	cmd.Run([]string{"-f", incomingFile, "-b", "123", "-c", "someFilter"})

	mockProvider.AssertExpectations(t)
}
