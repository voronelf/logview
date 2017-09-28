package settings

import (
	"github.com/pelletier/go-toml"
	"github.com/voronelf/logview/core"
	"io"
	"os"
	"path/filepath"
)

func NewStore() *store {
	home := os.Getenv("HOME")
	if home == "" {
		home = "~"
	}
	return &store{
		filePath: filepath.Join(home, ".logview", "settings.toml"),
	}
}

type store struct {
	filePath string
	fileName string
}

var _ core.Settings = (*store)(nil)

func (s *store) GetTemplates() (map[string]core.Template, error) {
	tree, err := toml.LoadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	content := tomlContent{}
	err = tree.Unmarshal(&content)
	if err != nil {
		return nil, err
	}
	result := map[string]core.Template{}
	for k, v := range content.Templates {
		result[k] = core.Template(v)
	}
	return result, nil
}

func (s *store) SaveTemplate(name string, tpl core.Template) error {
	dir := filepath.Dir(s.filePath)
	if _, e := os.Stat(dir); os.IsNotExist(e) {
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			return err
		}
	}
	fd, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0775)
	if err != nil {
		return err
	}
	defer fd.Close()

	tree, err := toml.LoadReader(fd)
	if err != nil {
		return err
	}

	for flag, value := range tpl {
		tree.Set("templates."+name+"."+flag, value)
	}

	_, err = fd.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	err = fd.Truncate(0)
	if err != nil {
		return err
	}
	_, err = tree.WriteTo(fd)
	if err != nil {
		return err
	}
	return nil
}

type tomlContent struct {
	Templates map[string]map[string]string `toml:"templates"`
}
