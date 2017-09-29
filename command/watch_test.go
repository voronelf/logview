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
	"strings"
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
	formatParams := core.DefaultFormatParams()
	formatParams.OutputFields = []string{"field1", "field2", "field3"}
	formatParams.AccentFields = []string{"field1", "field3"}
	mockFilter := &core.MockFilter{}
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, "someFile").Return((<-chan core.Row)(rowsChan), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row, formatParams).Return("SomeData").Twice()

	go cmd.Run([]string{"-f", "someFile", "-c", "someFilter", "-o", "field1,field2,field3", "-a", "field1,field3"})
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
	formatParams := core.DefaultFormatParams()
	formatParams.OutputFields = []string{"field1", "field2", "field3"}
	formatParams.AccentFields = []string{"field1", "field3"}
	mockFilter := &core.MockFilter{}
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchOpenedStream", mock.Anything, cmd.Stdin).Return((<-chan core.Row)(rowsCh), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row, formatParams).Return("SomeData").Twice()

	go cmd.Run([]string{"-c", "someFilter", "-o", "field1,field2,field3", "-a", "field1,field3"})
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
	templates := map[string]core.Template{
		"someTpl": {"f": "someFile", "c": "someFilter", "o": "field1,field2,field3", "a": "field1,field3"},
	}
	formatParams := core.DefaultFormatParams()
	formatParams.OutputFields = []string{"field1", "field2", "field3"}
	formatParams.AccentFields = []string{"field1", "field3"}
	mockSettings.On("GetTemplates").Return(templates, nil)
	mockFilterFactory.On("NewFilter", "someFilter").Return(mockFilter, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, "someFile").Return((<-chan core.Row)(rowsChan), nil).Once()
	mockFilter.On("Match", row).Return(true).Twice()
	mockFormatter.On("Format", row, formatParams).Return("SomeData").Twice()

	go cmd.Run([]string{"-t", "someTpl"})
	time.Sleep(2 * time.Millisecond)
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
	prmsDefault := core.DefaultFormatParams()
	tplSet_1 := map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond"}}
	tplSet_2 := map[string]core.Template{"tpl1": {"f": "tplFile", "c": "tplCond", "o": "field1,field2,field3", "a": "field1,field3"}}
	prms_2 := core.DefaultFormatParams()
	prms_2.OutputFields = []string{"field1", "field2", "field3"}
	prms_2.AccentFields = []string{"field1", "field3"}
	cases := []struct {
		args   string
		tpls   map[string]core.Template
		file   string
		cond   string
		params core.FormatParams
		err    bool
	}{
		{"-f someFile -c someCond", map[string]core.Template{}, "someFile", "someCond", prmsDefault, false},
		{"-f someFile -c someCond", tplSet_1, "someFile", "someCond", prmsDefault, false},
		{"-f someFile -c someCond -t tpl1", tplSet_1, "someFile", "someCond", prmsDefault, false},
		{"-f someFile -t tpl1", tplSet_1, "someFile", "tplCond", prmsDefault, false},
		{"-c someCond -t tpl1", tplSet_1, "tplFile", "someCond", prmsDefault, false},
		{"-t tpl1", tplSet_1, "tplFile", "tplCond", prmsDefault, false},
		{"-t tpl2", tplSet_1, "", "", prmsDefault, true},
		{"-f someFile -c someCond -o field1,field2,field3 -a field1,field3", map[string]core.Template{}, "someFile", "someCond", prms_2, false},
		{"-t tpl1", tplSet_2, "tplFile", "tplCond", prms_2, false},
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
			actualFilePath, actualCondition, params, err := cmd.parseArgs(strings.Split(cs.args, " "))
			assert.Equal(t, cs.file, actualFilePath)
			assert.Equal(t, cs.cond, actualCondition)
			assert.Equal(t, cs.params, params)
			if cs.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestWatch_Run_FileWithDate(t *testing.T) {
	cmd, shutdownCh := newWatchForTest()
	defer close(shutdownCh)
	mockProvider := cmd.RowProvider.(*core.MockRowProvider)
	mockFilterFactory := cmd.FilterFactory.(*core.MockFilterFactory)

	incomingFile := "someFile_@today@.log"
	expectedFile := "someFile_" + time.Now().UTC().Format("2006-01-02") + ".log"

	mockFilterFactory.On("NewFilter", "someFilter").Return(&core.MockFilter{}, nil).Once()
	mockProvider.On("WatchFileChanges", mock.Anything, expectedFile).Return(make(<-chan core.Row), nil).Once()

	go cmd.Run([]string{"-f", incomingFile, "-c", "someFilter"})
	time.Sleep(time.Millisecond)

	mockProvider.AssertExpectations(t)
}
