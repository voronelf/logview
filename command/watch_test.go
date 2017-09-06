package command

import (
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
		Observer:      &core.MockObserver{},
		FilterFactory: &core.MockFilterFactory{},
		Formatter:     &core.MockFormatter{},
		Ui:            &cli.MockUi{},
		ShutdownCh:    shutdownCh,
	}, shutdownCh
}

func TestWatch_Run(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockObserver := cmd.Observer.(*core.MockObserver)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	filter := &core.MockFilter{}
	inputCh := make(chan core.Row, 1)
	s := core.Subscription{Channel: inputCh}
	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}

	mockFilterFactory.On("NewFilter", "someFilter").Return(filter, nil).Once()
	mockObserver.On("Subscribe", mock.Anything, "someFile", filter).Return(s, nil).Once()
	mockFormatter.On("Format", row).Return("SomeData").Twice()

	go cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})
	time.Sleep(time.Millisecond)
	inputCh <- row
	time.Sleep(time.Millisecond)
	inputCh <- row
	time.Sleep(time.Millisecond)

	mockFilterFactory.AssertExpectations(t)
	mockObserver.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	expectedOutput := "Watch file someFile with filter \"someFilter\"\nSomeData\nSomeData\n"
	assert.Equal(t, expectedOutput, cmd.Ui.(*cli.MockUi).OutputWriter.String())
}

func TestWatch_Run_Shutdown(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()

	cmd.FilterFactory.(*core.MockFilterFactory).On("NewFilter", mock.Anything).Return(&core.MockFilter{}, nil)
	s := core.Subscription{Channel: make(chan core.Row, 1)}
	cmd.Observer.(*core.MockObserver).On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(s, nil)

	done := false
	mu := sync.Mutex{}
	go func() {
		cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})
		mu.Lock()
		done = true
		mu.Unlock()
	}()

	close(shutdownCh)
	time.Sleep(5 * time.Millisecond)
	mu.Lock()
	assert.True(t, done)
	mu.Unlock()
}

func TestWatch_Run_SubscribeError(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockObserver := cmd.Observer.(*core.MockObserver)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	filter := &core.MockFilter{}

	mockFilterFactory.On("NewFilter", "someFilter").Return(filter, nil).Once()
	mockObserver.On("Subscribe", mock.Anything, "someFile", filter).Return(core.Subscription{}, errors.New("Some error")).Once()

	cmd.Run([]string{"-f", "someFile", "-c", "someFilter"})

	mockFilterFactory.AssertExpectations(t)
	mockObserver.AssertExpectations(t)
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
