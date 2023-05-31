package worker

import (
	"context"
	"log"
	"sync"
)

type Result struct {
	Value   any
	Err     error
	Context context.Context
}

type Job struct {
	Fn func(ctx context.Context) (any, error)
}

func (job *Job) Do(ctx context.Context) Result {
	value, err := job.Fn(ctx)
	if err != nil {
		return Result{
			Err:     err,
			Context: ctx,
		}
	}

	return Result{
		Value:   value,
		Context: ctx,
	}
}

type Pool struct {
	workerCount int
	jobs        <-chan Job
}

func NewPool(workerCount int) *Pool {
	return &Pool{
		workerCount: workerCount,
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}

			results <- job.Do(ctx)
		case <-ctx.Done():
			log.Println(ctx.Err())
		}
	}
}

func (pool *Pool) Run(ctx context.Context) <-chan Result {
	results := make(chan Result, pool.workerCount)

	wg := new(sync.WaitGroup)
	wg.Add(pool.workerCount)
	for i := 0; i < pool.workerCount; i++ {
		go worker(ctx, wg, pool.jobs, results)
	}

	go func() {
		wg.Wait()

		close(results)
	}()

	return results
}

func (pool *Pool) AddJobs(jobs <-chan Job) {
	pool.jobs = jobs
}
