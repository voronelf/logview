package core

import "context"

type Row struct {
	Data map[string]interface{}
	Err  error
}

type Subscription struct {
	Channel <-chan Row
}

//go:generate mockery -name Observer -inpkg -case=underscore

type Observer interface {
	Subscribe(ctx context.Context, filePath string, filter Filter) (Subscription, error)
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
	Format(row Row) string
}

//go:generate mockery -name FileReader -inpkg -case=underscore

type FileReader interface {
	ReadTail(filePath string, b int64, filter Filter) (<-chan Row, error)
}
