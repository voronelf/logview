package core

import (
	"context"
	"io"
)

type Row struct {
	Data map[string]interface{}
	Err  error
}

type Subscription struct {
	Channel <-chan Row
}

//go:generate mockery -name RowProvider -inpkg -case=underscore

type RowProvider interface {
	WatchFileChanges(ctx context.Context, filePath string) (<-chan Row, error)
	WatchOpenedStream(ctx context.Context, stream io.Reader) (<-chan Row, error)
	ReadFileTail(ctx context.Context, filePath string, countBytes int64) (<-chan Row, error)
}

//go:generate mockery -name Filter -inpkg -case=underscore

type Filter interface {
	Match(row Row) bool
}

//go:generate mockery -name FilterFactory -inpkg -case=underscore

type FilterFactory interface {
	NewFilter(condition string) (Filter, error)
}

//go:generate mockery -name Formatter -inpkg -case=underscore

type Formatter interface {
	Format(row Row, params FormatParams) string
}

type FormatParams struct {
	OutputFields []string
	AccentFields []string
}

func DefaultFormatParams() FormatParams {
	return FormatParams{
		OutputFields: []string{},
		AccentFields: []string{},
	}
}

//go:generate mockery -name Settings -inpkg -case=underscore

type Settings interface {
	GetTemplates() (map[string]Template, error)
	SaveTemplate(name string, tpl Template) error
}

type Template map[string]string
