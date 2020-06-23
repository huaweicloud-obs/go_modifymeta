package limiter

import (
	"sync/atomic"
)

const (
	//DefaultLimit the default maximum limit of goroutine
	DefaultLimit = 2000
)

//ConcurrencyLimiter a limiter of goroutine
type ConcurrencyLimiter struct {
	Limit         int      `json:"limit"`
	Tickets       chan int `json:"tickets"`
	NumInProgress int32    `json:"in_progress"`
}

//NewConcurrencyLimiter enforce a maximum Concurrency of limit
func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	if limit <= 0 {
		limit = DefaultLimit
	}

	// allocate a limiter instance
	c := &ConcurrencyLimiter{
		Limit:   limit,
		Tickets: make(chan int, limit),
	}

	// allocate the tickets:
	for i := 0; i < c.Limit; i++ {
		c.Tickets <- i
	}

	return c
}

// Execute launch a new routine to execute job
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConcurrencyLimiter) Execute(job func()) int {
	ticket := <-c.Tickets
	//fmt.Println("now total pid is:",ticket)
	atomic.AddInt32(&c.NumInProgress, 1)
	go func() {
		defer func() {
			c.Tickets <- ticket
			atomic.AddInt32(&c.NumInProgress, -1)

		}()

		// run the job
		job()
	}()
	return ticket
}

// ExecuteWithParams launch a new routine to execute job
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConcurrencyLimiter) ExecuteWithParams(job func(args ...interface{}), jobParams ...interface{}) int {
	ticket := <-c.Tickets
	atomic.AddInt32(&c.NumInProgress, 1)
	go func() {
		defer func() {
			c.Tickets <- ticket
			atomic.AddInt32(&c.NumInProgress, -1)
		}()

		// run the job
		job(jobParams...)
	}()
	return ticket
}

// ExecuteWithTicket launch a new go routine with ticket to execute job
// if num of go routines allocated by this instance is < limit
// launch a new go routine to execute job
// else wait until a go routine becomes available
func (c *ConcurrencyLimiter) ExecuteWithTicket(job func(ticket int)) int {
	ticket := <-c.Tickets
	atomic.AddInt32(&c.NumInProgress, 1)
	go func() {
		defer func() {
			c.Tickets <- ticket
			atomic.AddInt32(&c.NumInProgress, -1)
		}()

		// run the job
		job(ticket)
	}()
	return ticket
}

// Wait wait until all the previously Executed jobs completed running
//
// IMPORTANT: calling the Wait function while keep calling Execute leads to
//            un-desired race conditions
func (c *ConcurrencyLimiter) Wait() {
	for i := 0; i < c.Limit; i++ {
		_ = <-c.Tickets
	}
}

// GetNumInProgress get a racy counter of how many go routines are active right now
func (c *ConcurrencyLimiter) GetNumInProgress() int32 {
	return c.NumInProgress
}
