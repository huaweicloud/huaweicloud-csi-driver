package services

import (
	"testing"
	"time"
)

func TestLockServerSerialisesOneServer(t *testing.T) {
	unlock := lockServer("ecs-a")
	got := make(chan struct{})
	go func() {
		defer lockServer("ecs-a")()
		close(got)
	}()

	select {
	case <-got:
		t.Fatal("a second attach to the same server ran while the first held the lock")
	case <-time.After(50 * time.Millisecond):
	}
	unlock()
	select {
	case <-got:
	case <-time.After(time.Second):
		t.Fatal("the second attach did not run after the first released the lock")
	}
}

func TestLockServerOtherServersDoNotWait(t *testing.T) {
	defer lockServer("ecs-b")()
	done := make(chan struct{})
	go func() {
		defer lockServer("ecs-c")()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("an attach to another server waited for an unrelated lock")
	}
}
