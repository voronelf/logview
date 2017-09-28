package settings

import (
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	home := filepath.Join("home", "someUser")
	os.Setenv("HOME", home)
	s := NewStore()
	assert.Equal(t, filepath.Join(home, ".logview", "settings.toml"), s.filePath)
}

func TestStore_GetTemplates(t *testing.T) {
	s := NewStore()
	s.filePath = "test/settings.toml"
	actual, err := s.GetTemplates()
	assert.Nil(t, err)
	assert.Equal(t, map[string]core.Template{"tpl1": {"f": "fff", "c": "ccc"}}, actual)
}

func TestStore_SaveTemplate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "logview_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	filePath := filepath.Join(tempDir, "subdir", "settings.toml")
	tpl := map[string]string{"f": "fff", "c": "ccc"}

	s := NewStore()
	s.filePath = filePath
	s.SaveTemplate("tpl1", tpl)

	tree, err := toml.LoadFile(filePath)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	actual := tomlContent{Templates: make(map[string]map[string]string)}
	err = tree.Unmarshal(&actual)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	assert.Equal(t, tomlContent{Templates: map[string]map[string]string{"tpl1": tpl}}, actual)
}
