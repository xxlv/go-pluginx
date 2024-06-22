package build

import "sync"

// TODO:  impl queue

type BuildQueue struct {
	queue []string
	mu    sync.Mutex
}

func (sq *BuildQueue) Push(v string) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.queue = append(sq.queue, v)
}
func (sq *BuildQueue) Log(v string) {
	sq.Push("log: " + v)
}

func (sq *BuildQueue) Error(v string) {
	sq.Push("error: " + v)
}

func (sq *BuildQueue) Pop() (string, bool) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	if len(sq.queue) == 0 {
		return "", false
	}
	val := sq.queue[0]
	sq.queue = sq.queue[1:]
	return val, true
}
