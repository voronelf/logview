package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type t1 struct{}

type t2 struct {
	Field *t1 `inject:"t1_name"`
}

func TestGetDIContainer(t *testing.T) {
	first := t1{}
	func() {
		di1 := GetDIContainer()
		di1.Provide("t1_name", &first)
	}()
	second := t2{}
	func() {
		di2 := GetDIContainer()
		di2.Provide("", &second)
	}()
	di3 := GetDIContainer()
	err := di3.Populate()
	assert.Nil(t, err)
	assert.Equal(t, &first, second.Field)
}

func TestDI_Populate(t *testing.T) {
	di := newDI()
	first := t1{}
	di.Provide("t1_name", &first)
	second := t2{}
	di.Provide("", &second)
	err := di.Populate()
	assert.Nil(t, err)
	assert.Equal(t, &first, second.Field)
}
