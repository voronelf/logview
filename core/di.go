package core

import (
	"errors"
	"github.com/facebookgo/inject"
)

var diContainer *DIContainer

// GetDIContainer returns global DIContainer container instance
func GetDIContainer() *DIContainer {
	if diContainer == nil {
		diContainer = newDI()
	}
	return diContainer
}

func newDI() *DIContainer {
	return &DIContainer{objects: []*inject.Object{}}
}

type DIContainer struct {
	objects []*inject.Object
}

func (di *DIContainer) Provide(name string, value interface{}) {
	di.objects = append(di.objects, &inject.Object{Value: value, Name: name})
}

func (di *DIContainer) Populate() error {
	var graph inject.Graph
	err := graph.Provide(di.objects...)
	if err != nil {
		return errors.New("inject.Graph.Provide() error: " + err.Error())
	}
	err = graph.Populate()
	if err != nil {
		return errors.New("inject.Graph.Populate() error: " + err.Error())
	}
	return nil
}
