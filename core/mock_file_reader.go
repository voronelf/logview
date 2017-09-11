package core

import mock "github.com/stretchr/testify/mock"

// MockFileReader is an autogenerated mock type for the FileReader type
type MockFileReader struct {
	mock.Mock
}

// ReadTail provides a mock function with given fields: filePath, b, filter
func (_m *MockFileReader) ReadTail(filePath string, b int64, filter Filter) (<-chan Row, error) {
	ret := _m.Called(filePath, b, filter)

	var r0 <-chan Row
	if rf, ok := ret.Get(0).(func(string, int64, Filter) <-chan Row); ok {
		r0 = rf(filePath, b, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan Row)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int64, Filter) error); ok {
		r1 = rf(filePath, b, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

var _ FileReader = (*MockFileReader)(nil)
