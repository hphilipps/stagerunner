package domain

import (
	"container/list"
	"fmt"
	"sync"
)

// queue implements a FIFO queue for pipeline runs which ensures we do not queue
// more than queueSize runs.
// Also ensures we do not queue more than maxQueuedPerPipeline runs per pipeline -
// this is to prevent a single pipeline to block the queue for all other pipelines.
type queue struct {
	queueSize            int
	maxQueuedPerPipeline int
	queue                *list.List
	pipelineCounts       map[string]int
	mu                   sync.Mutex
}

// newQueue creates a new queue with the given size and max queued per pipeline.
func newQueue(queueSize, maxQueuedPerPipeline int) *queue {
	return &queue{
		queueSize:            queueSize,
		maxQueuedPerPipeline: maxQueuedPerPipeline,
		queue:                list.New(),
		pipelineCounts:       make(map[string]int),
	}
}

func (q *queue) Enqueue(item *PipelineRun) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if the overall queue is full
	if q.queue.Len() >= q.queueSize {
		return fmt.Errorf("queue is full - can not enqueue more than %d runs, consider using more workers", q.queueSize)
	}

	// Check if the pipeline has exceeded its maximum queued runs
	if q.pipelineCounts[item.PipelineID] >= q.maxQueuedPerPipeline {
		return fmt.Errorf("pipeline %s has reached its maximum queued runs - can not enqueue more than %d runs per pipeline", item.PipelineID, q.maxQueuedPerPipeline)
	}

	// Add item to the queue
	q.queue.PushBack(item)
	q.pipelineCounts[item.PipelineID]++
	return nil
}

func (q *queue) Dequeue() (*PipelineRun, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue.Len() == 0 {
		return nil, ErrQueueEmpty
	}

	element := q.queue.Front()
	q.queue.Remove(element)

	item := element.Value.(*PipelineRun)
	q.pipelineCounts[item.PipelineID]--

	// Remove pipeline entry if count is zero
	if q.pipelineCounts[item.PipelineID] == 0 {
		delete(q.pipelineCounts, item.PipelineID)
	}

	return item, nil
}
