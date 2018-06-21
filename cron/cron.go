package cron

import (
	"fmt"
	"runtime"
	"sort"
	"time"
)

// 提供删除job的检查函数，返回true则删除
type RemoveCheckFunc func(e *Entry) bool

// cron调度
type Cron struct {
	entries  []*Entry
	add      chan *Entry
	remove   chan RemoveCheckFunc
	stop     chan bool
	running  bool
	snapshot chan []*Entry
}

// Job is an interface for submitted cron jobs.
type Job interface {
	Run()
}

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	Name string

	// shedule the job
	Schedule *Schedule

	// the next time the job will run
	Next time.Time

	// the last time this job was run
	Prev time.Time

	// the job to run
	Job Job
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

func New() *Cron {
	return &Cron{
		entries:  nil,
		add:      make(chan *Entry),
		remove:   make(chan RemoveCheckFunc),
		stop:     make(chan bool),
		snapshot: make(chan []*Entry),
		running:  false,
	}
}

// A wrapper that turns a func() into a cron.Job
type FuncJob func()

func (f FuncJob) Run() { f() }

// AddFunc adds a func to the Cron to be run on the schedule
func (c *Cron) AddFunc(name, cronExpr string, cmd func()) error {
	return c.AddJob(name, cronExpr, FuncJob(cmd))
}

func (c *Cron) AddJob(name, cronExpr string, cmd Job) error {
	schedule, err := Parse(cronExpr)
	if err != nil {
		return err
	}
	c.Schedule(name, schedule, cmd)
	return nil
}

// RemoveJob remove a job from the cron
func (c *Cron) RemoveJob(cb RemoveCheckFunc) {
	if c.running {
		c.remove <- cb
	} else {
		c.removeJob(cb)
	}
}

func (c *Cron) removeJob(cb RemoveCheckFunc) {
	newEntries := make([]*Entry, 0)
	for _, e := range c.entries {
		if !cb(e) {
			newEntries = append(newEntries, e)
		}
	}
	c.entries = newEntries
}

// Schedule adds a Job to the Cron to be run on the given shedule.
func (c *Cron) Schedule(name string, schedule *Schedule, cmd Job) {
	entry := &Entry{
		Name:     name,
		Schedule: schedule,
		Job:      cmd,
	}
	if !c.running {
		c.entries = append(c.entries, entry)
		return
	}
	c.add <- entry
}

// Entries returns a snapshot of the cron entries.
func (c *Cron) Entries() []*Entry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

func (c *Cron) Start() {
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

func (c *Cron) Stop() {
	if !c.running {
		return
	}
	c.stop <- true
	c.running = false
}

func (c *Cron) IsRunning() bool {
	return c.running
}

func (c *Cron) runWithRecovery(e *Entry) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err := fmt.Sprintf(
				" error: panic running job[start at %s]: %v\n%s",
				start.Format("2006/01/02 15:04:05"),
				r,
				buf,
			)
			log.Error(e.Name, err)
		}
	}()
	e.Job.Run()
	end := time.Now()
	latency := end.Sub(start)
	info := fmt.Sprintf(
		"start: %s  end: %s  latency: %sms",
		start.Format("2006/01/02 15:04:05"),
		end.Format("2006/01/02 15:04:05"),
		latency.Nanoseconds()/1000,
	)
	log.Info(e.Name, info)
}

func (c *Cron) run() {
	now := time.Now()
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			timer = time.NewTimer(c.entries[0].Next.Sub(now))
		}

		select {
		case now = <-timer.C:
			for _, e := range c.entries {
				if e.Next.After(now) || e.Next.IsZero() {
					break
				}
				log.Info(e.Name, " is ready to start at ", now.String())
				go c.runWithRecovery(e)
				e.Prev = e.Next
				e.Next = e.Schedule.Next(now)
			}
		case newEntry := <-c.add:
			timer.Stop()
			now = time.Now()
			newEntry.Next = newEntry.Schedule.Next(now)
			c.entries = append(c.entries, newEntry)
		case cb := <-c.remove:
			c.removeJob(cb)

		case <-c.snapshot:
			c.snapshot <- c.entrySnapshot()
			continue

		case <-c.stop:
			timer.Stop()
			return
		}
	}
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *Cron) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Name:     e.Name,
			Schedule: e.Schedule,
			Next:     e.Next,
			Prev:     e.Prev,
			Job:      e.Job,
		})
	}
	return entries
}
