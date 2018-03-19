package worker

import (
    "context"
    "log"

    "golang.org/x/sync/semaphore"
)

type Worker struct {
    sem *semaphore.Weighted
    ctx context.Context
    numWorker int64
}

func NewWorker (numWorker int64) Worker {
    return Worker{semaphore.NewWeighted(numWorker), context.TODO(), numWorker}
}

func (w *Worker) acquire (num int64) int {
    if err := w.sem.Acquire(w.ctx, num); err != nil {
        log.Printf("Worker: failed to acquire semaphore: %v", err)
        return -1
    }

    return 0
}

func (w *Worker) Spawn (f func()) int {
    if ret := w.acquire(1); ret < 0 {
        return ret
    }

    go func() {
        defer w.sem.Release(1)
        f()
    }()

    return 0
}

func (w *Worker) WaitAll () int {
    if ret := w.acquire(w.numWorker); ret < 0 {
        return ret
    }

    return 0
}
