package session

import (
	"sync"
	"testing"
	"time"
)

func TestPipeClose(t *testing.T) {
	t.Parallel()

	local, remote := Pipe(time.Second)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		select {
		case _, ok := <-local.In:
			if ok {
				t.Error("local read without send")
			}
		case <-time.After(time.Second):
			t.Error("local inbound channel close timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case _, ok := <-remote.In:
			if ok {
				t.Error("remote read without send")
			}
		case <-time.After(time.Second):
			t.Error("remote inbound channel close timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case err, ok := <-local.Err:
			if ok {
				t.Error("local error:", err)
			}
		case <-time.After(time.Second):
			t.Error("local error channel close timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case err, ok := <-remote.Err:
			if ok {
				t.Error("remote error:", err)
			}
		case <-time.After(time.Second):
			t.Error("remote error channel close timeout")
		}
	}()

	// local close must shut down both ends
	close(local.Class1)
	close(local.Class2)
	wg.Wait()

	close(remote.Class1)
	close(remote.Class2)
}
