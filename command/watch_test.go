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
		Settings:      &core.MockSettings{},
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

func TestWatch_Run_Template(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockSettings := cmd.Settings.(*core.MockSettings)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFormatter := cmd.Formatter.(*core.MockFormatter)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	rowsChan := make(chan core.Row, 2)
	row := core.Row{Data: map[string]interface{}{"someKey": "someValue"}}
	mockFilter := &core.MockFilter{}
	mockSettings.On("GetTemplates").Return(map[string]core.Template{"someTpl": {"f": "someFile", "c": "someFilter"}}, nil)
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, "someFile").Return((<-chan core.Row)(rowsChan), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row).Return("SomeData").Twice()

	go cmd.Run([]string{"-t", "someTpl"})
	time.Sleep(time.Millisecond)
	rowsChan <- row
	time.Sleep(time.Millisecond)
	rowsChan <- row
	time.Sleep(time.Millisecond)

	mockSettings.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockFilterFactory.AssertExpectations(t)
	mockFilter.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
	expectedOutput := messageWatchFile("someFile", "someFilter") + "\nSomeData\nSomeData\n"
	assert.Equal(t, expectedOutput, cmd.Ui.(*cli.MockUi).OutputWriter.String())
}

func TestWatch_parseArgs(t *testing.T) {
	cases := []struct {
		args string
		tpls map[string]core.Template
		file string
		cond string
		err  bool
	}{
		{"-f someFile -c someCond", map[string]core.Template{}, "someFile", "someCond", false},
		{"-f someFile -c someCond", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "someFile", "someCond", false},
		{"-f someFile -c someCond -t tpl1", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "someFile", "someCond", false},
		{"-f someFile -t tpl1", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "someFile", "tplCond", false},
		{"-c someCond -t tpl1", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "tplFile", "someCond", false},
		{"-t tpl1", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "tplFile", "tplCond", false},
		{"-t tpl2", map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}, "", "", true},
	}
	for i, cs := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cmd, shutdownCh := newWatchForTest()
			defer close(shutdownCh)
			var retErr error
			if cs.err {
				retErr = errors.New("some err")
			}
			cmd.Settings.(*core.MockSettings).On("GetTemplates").Return(cs.tpls, retErr)
			actualFilePath, actualCondition, err := cmd.parseArgs(strings.Split(cs.args, " "))
			assert.Equal(t, cs.file, actualFilePath)
			assert.Equal(t, cs.cond, actualCondition)
			if cs.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
