package session

import (
	"testing"
	"time"
)

func TestLoginGuardLimitsAndResetsAttempts(t *testing.T) {
	guard := newLoginGuard()
	now := time.Unix(100, 0)
	guard.now = func() time.Time { return now }
	for attempt := 0; attempt < loginAttemptsPerUser; attempt++ {
		if !guard.allow("127.0.0.1", "Tester") {
			t.Fatalf("attempt %d was unexpectedly rejected", attempt+1)
		}
	}
	if guard.allow("127.0.0.2", "tester") {
		t.Fatal("username limit was not enforced case-insensitively")
	}
	now = now.Add(loginWindow)
	if !guard.allow("127.0.0.1", "tester") {
		t.Fatal("attempt window did not reset")
	}
}

func TestLoginGuardBoundsConcurrentPasswordChecks(t *testing.T) {
	guard := newLoginGuard()
	for index := 0; index < maxPasswordChecks; index++ {
		if !guard.acquire() {
			t.Fatalf("slot %d was unexpectedly rejected", index+1)
		}
	}
	if guard.acquire() {
		t.Fatal("concurrency limit was not enforced")
	}
	guard.release()
	if !guard.acquire() {
		t.Fatal("released concurrency slot was not reusable")
	}
}
