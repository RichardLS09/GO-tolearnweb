package cron

import (
	"sync"
	"testing"
	"time"
)

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const OneMinute = 1*time.Minute + 10*time.Millisecond
const OneSecond = 1*time.Second + 10*time.Millisecond

func TestFuncPanicRecovery(t *testing.T) {
	cron := New()
	cron.Start()
	defer cron.Stop()
	cron.AddFunc("YOLO", "* * * * *", func() { panic("YOLO") })
	select {
	case <-time.After(OneSecond):
		return
	}
}

type DummyJob struct{}

func (d DummyJob) Run() {
	panic("YOLO")
}

// Start and stop cron with no entries.
func TestNoEntries(t *testing.T) {
	cron := New()
	cron.Start()

	select {
	case <-time.After(OneSecond):
		t.Fatal("expected cron will be stopped immediately")
	case <-stop(cron):
	}
}

// Test that calling stop before start silently returns without
// blocking the stop channel.
func TestStopWithoutStart(t *testing.T) {
	cron := New()
	cron.Stop()
}

type testJob struct {
	wg   *sync.WaitGroup
	name string
}

func (t testJob) Run() {
	t.wg.Done()
}

// Test that adding an invalid job spec returns an error
func TestInvalidJobSpec(t *testing.T) {
	cron := New()
	err := cron.AddJob("wrong", "this will not parse", nil)
	if err == nil {
		t.Errorf("expected an error with invalid spec, got nil")
	}
}

func TestJob(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	if err := cron.AddJob("job0", "0 0 30 2 *", testJob{wg, "job0"}); err != nil {
		t.Error(err)
	}
	if err := cron.AddJob("job1", "0 0 1 1 *", testJob{wg, "job1"}); err != nil {
		t.Error(err)
	}
	if err := cron.AddJob("job2", "* * * * *", testJob{wg, "job2"}); err != nil {
		t.Error(err)
	}
	if err := cron.AddJob("job3", "1 0 1 1 *", testJob{wg, "job3"}); err != nil {
		t.Error(err)
	}

	cron.Start()
	defer cron.Stop()

	select {
	case <-time.After(OneMinute):
		t.FailNow()
	case <-wait(wg):
	}

	// Ensure the entries are in the right order.
	expecteds := []string{"job2", "job1", "job3", "job0"}

	var actuals []string
	for _, entry := range cron.Entries() {
		actuals = append(actuals, entry.Name)
	}

	for i, expected := range expecteds {
		if actuals[i] != expected {
			t.Fatalf("Jobs not in the right order.  (expected) %s != %s (actual)", expecteds, actuals)
		}
	}
}

// // Add a job, start cron, expect it runs.
// func TestAddBeforeRunning(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)

// 	cron := New()
// 	cron.AddFunc("add before running", "* * * * *", func() { wg.Done() })
// 	cron.Start()
// 	defer cron.Stop()

// 	// Give cron 2 seconds to run our job (which is always activated).
// 	select {
// 	case <-time.After(OneMinute):
// 		t.Fatal("expected job runs")
// 	case <-wait(wg):
// 	}
// }

// // Start cron, add a job, expect it runs.
// func TestAddWhileRunning(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)

// 	cron := New()
// 	cron.Start()
// 	defer cron.Stop()
// 	cron.AddFunc("add while running", "* * * * *", func() { wg.Done() })

// 	select {
// 	case <-time.After(OneMinute):
// 		t.Fatal("expected job runs")
// 	case <-wait(wg):
// 	}
// }

// // Test for #34. Adding a job after calling start results in multiple job invocations
// func TestAddWhileRunningWithDelay(t *testing.T) {
// 	cron := New()
// 	cron.Start()
// 	defer cron.Stop()
// 	time.Sleep(5 * time.Second)
// 	var calls = 0
// 	cron.AddFunc("add while running with delay", "* * * * *", func() { calls += 1 })

// 	<-time.After(OneMinute)
// 	if calls != 1 {
// 		t.Errorf("called %d times, expected 1\n", calls)
// 	}
// }

// // Test that the entries are correctly sorted.
// // Add a bunch of long-in-the-future entries, and an immediate entry, and ensure
// // that the immediate entry runs immediately.
// // Also: Test that multiple jobs run in the same instant.
// func TestMultipleEntries(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(2)

// 	cron := New()
// 	cron.AddFunc("null", "0 0 1 1 *", func() {})
// 	cron.AddFunc("done", "* * * * *", func() { wg.Done() })
// 	cron.AddFunc("null2", "0 0 31 12 *", func() {})
// 	cron.AddFunc("done2", "* * * * *", func() { wg.Done() })

// 	cron.Start()
// 	defer cron.Stop()

// 	select {
// 	case <-time.After(OneMinute):
// 		t.Error("expected job run in proper order")
// 	case <-wait(wg):
// 	}
// }

// // Test running the same job twice.
// func TestRunningJobTwice(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(2)

// 	cron := New()
// 	cron.AddFunc("null", "0 0 1 1 *", func() {})
// 	cron.AddFunc("null2", "0 0 31 12 *", func() {})
// 	cron.AddFunc("done", "* * * * *", func() { wg.Done() })

// 	cron.Start()
// 	defer cron.Stop()

// 	select {
// 	case <-time.After(2 * OneMinute):
// 		t.Error("expected job fires 2 times")
// 	case <-wait(wg):
// 	}
// }

// func TestRunningMultipleSchedules(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(2)

// 	cron := New()
// 	cron.AddFunc("null", "0 0 1 1 *", func() {})
// 	cron.AddFunc("null2", "0 0 31 12 *", func() {})
// 	cron.AddFunc("done", "* * * * *", func() { wg.Done() })

// 	cron.Start()
// 	defer cron.Stop()

// 	select {
// 	case <-time.After(2 * OneMinute):
// 		t.Error("expected job fires 2 times")
// 	case <-wait(wg):
// 	}
// }

// func TestJobPanicRecovery(t *testing.T) {
// 	var job DummyJob

// 	cron := New()
// 	cron.Start()
// 	defer cron.Stop()
// 	cron.AddJob("panic recovery", "* * * * *", job)

// 	select {
// 	case <-time.After(OneMinute):
// 		return
// 	}
// }

// // Test that double-running is a no-op
// func TestStartNoop(t *testing.T) {
// 	var tickChan = make(chan struct{}, 2)

// 	cron := New()
// 	cron.AddFunc("start no-op", "* * * * *", func() {
// 		tickChan <- struct{}{}
// 	})

// 	cron.Start()
// 	defer cron.Stop()

// 	// Wait for the first firing to ensure the runner is going
// 	<-tickChan

// 	cron.Start()

// 	<-tickChan

// 	// Fail if this job fires again in a short period, indicating a double-run
// 	select {
// 	case <-time.After(time.Millisecond):
// 	case <-tickChan:
// 		t.Error("expected job fires exactly twice")
// 	}
// }

// // Start, stop, then add an entry. Verify entry doesn't run.
// func TestStopCausesJobsToNotRun(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)

// 	cron := New()
// 	cron.Start()
// 	cron.Stop()
// 	cron.AddFunc("stop jobs not run", "* * * * *", func() { wg.Done() })

// 	select {
// 	case <-time.After(OneSecond):
// 		// No job ran!
// 	case <-wait(wg):
// 		t.Fatal("expected stopped cron does not run any job")
// 	}
// }

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

func stop(cron *Cron) chan bool {
	ch := make(chan bool)
	go func() {
		cron.Stop()
		ch <- true
	}()
	return ch
}
