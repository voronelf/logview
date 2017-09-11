package command

import (
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func newTailForTest() (*Tail, chan<- struct{}) {
	shutdownCh := make(chan struct{})
	return &Tail{
		FileReader:    &core.MockFileReader{},
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
	mockFileReader := cmd.FileReader.(*core.MockFileReader)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)

	filter := &core.MockFilter{}
	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}
	channel := make(chan core.Row, 2)
	channel <- row
	channel <- row
	close(channel)
	mockFilterFactory.On("NewFilter", "someFilter").Return(filter, nil).Once()
	mockFileReader.On("ReadTail", "someFile", int64(123), filter).Return((<-chan core.Row)(channel), nil).Once()
	mockFormatter.On("Format", row).Return("SomeData").Twice()

	cmd.Run([]string{"-f", "someFile", "-b", "123", "-c", "someFilter"})
	mockFilterFactory.AssertExpectations(t)
	mockFileReader.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	assert.Equal(t, "SomeData\nSomeData\n", cmd.Ui.(*cli.MockUi).OutputWriter.String())
}
