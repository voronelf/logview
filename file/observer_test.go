package file

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/voronelf/logview/core"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

func TestObserverImpl_Subscribe(t *testing.T) {
	file, err := ioutil.TempFile("", "go_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()
	_, err = file.Write([]byte("{\"asdf\": 123}"))
	if err != nil {
		t.Fatal(err)
	}

	obs := NewObserver()
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	filter := &core.MockFilter{}
	filter.On("Match", mock.Anything).Return(true)
	subscription, err := obs.Subscribe(ctx, file.Name(), filter)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	catched := make([]core.Row, 0, 3)
	mu := sync.Mutex{}
	go func() {
		for {
			select {
			case row := <-subscription.Channel:
				mu.Lock()
				catched = append(catched, row)
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	time.Sleep(time.Millisecond)
	file.Write([]byte("{\"field\": \"456\"}\n"))
	time.Sleep(time.Millisecond)
	mu.Lock()
	if assert.Equal(t, 1, len(catched)) {
		assert.Equal(t, "456", catched[0].Data["field"])
		assert.Nil(t, catched[0].Err)
	}
	mu.Unlock()

	file.Write([]byte("{\"field\": 777}\n{\"field\": 888.99}\n"))
	time.Sleep(2 * time.Millisecond)
	mu.Lock()
	if assert.Equal(t, 3, len(catched)) {
		assert.Equal(t, float64(777), catched[1].Data["field"])
		assert.Nil(t, catched[1].Err)
		assert.Equal(t, float64(888.99), catched[2].Data["field"])
		assert.Nil(t, catched[2].Err)
	}
	mu.Unlock()
}
