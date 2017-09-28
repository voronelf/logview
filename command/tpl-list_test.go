package command

import (
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func TestTplList_Run(t *testing.T) {
	mockSettings := &core.MockSettings{}
	mockUi := &cli.MockUi{}
	cmd := &TplList{
		Settings: mockSettings,
		Ui:       mockUi,
	}
	mockSettings.On("GetTemplates").Return(map[string]core.Template{"tpl": {"f": "someFile"}}, nil)

	cmd.Run([]string{})
	mockSettings.AssertExpectations(t)

	assert.Contains(t, mockUi.OutputWriter.String(), "tpl")
	assert.Contains(t, mockUi.OutputWriter.String(), "f")
	assert.Contains(t, mockUi.OutputWriter.String(), "someFile")
}
