package domain

import (
	"testing"
)

func TestQueue_EnqueueDequeue(t *testing.T) {
	// Test basic queue operations
	t.Run("basic enqueue and dequeue", func(t *testing.T) {
		q := newQueue(5, 2)
		run := NewPipelineRun("pipeline1", "main")

		err := q.Enqueue(run)
		if err != nil {
			t.Errorf("unexpected error on enqueue: %v", err)
		}

		dequeued, err := q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error on dequeue: %v", err)
		}
		if dequeued != run {
			t.Error("dequeued item does not match enqueued item")
		}
	})

	// Test queue size limit
	t.Run("queue size limit", func(t *testing.T) {
		q := newQueue(2, 2)
		run1 := NewPipelineRun("pipeline1", "main")
		run2 := NewPipelineRun("pipeline2", "main")
		run3 := NewPipelineRun("pipeline3", "main")

		// Fill queue to capacity
		if err := q.Enqueue(run1); err != nil {
			t.Errorf("unexpected error on first enqueue: %v", err)
		}
		if err := q.Enqueue(run2); err != nil {
			t.Errorf("unexpected error on second enqueue: %v", err)
		}

		// Try to exceed capacity
		err := q.Enqueue(run3)
		if err == nil {
			t.Error("expected error when exceeding queue capacity")
		}
	})

	// Test per-pipeline limit
	t.Run("per-pipeline limit", func(t *testing.T) {
		q := newQueue(10, 2)
		run1 := NewPipelineRun("pipeline1", "main")
		run2 := NewPipelineRun("pipeline1", "dev")
		run3 := NewPipelineRun("pipeline1", "feature")

		// Fill pipeline to its limit
		if err := q.Enqueue(run1); err != nil {
			t.Errorf("unexpected error on first enqueue: %v", err)
		}
		if err := q.Enqueue(run2); err != nil {
			t.Errorf("unexpected error on second enqueue: %v", err)
		}

		// Try to exceed per-pipeline limit
		err := q.Enqueue(run3)
		if err == nil {
			t.Error("expected error when exceeding per-pipeline limit")
		}
	})

	// Test dequeue from empty queue
	t.Run("dequeue from empty queue", func(t *testing.T) {
		q := newQueue(5, 2)

		_, err := q.Dequeue()
		if err != ErrQueueEmpty {
			t.Errorf("expected ErrQueueEmpty, got %v", err)
		}
	})

	// Test pipeline count cleanup
	t.Run("pipeline count cleanup", func(t *testing.T) {
		q := newQueue(5, 2)
		run := NewPipelineRun("pipeline1", "main")

		if err := q.Enqueue(run); err != nil {
			t.Errorf("unexpected error on enqueue: %v", err)
		}

		// Verify pipeline count is tracked
		if q.pipelineCounts["pipeline1"] != 1 {
			t.Errorf("expected pipeline count 1, got %d", q.pipelineCounts["pipeline1"])
		}

		// Dequeue and verify cleanup
		_, err := q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error on dequeue: %v", err)
		}

		if count, exists := q.pipelineCounts["pipeline1"]; exists {
			t.Errorf("expected pipeline count to be cleaned up, but got %d", count)
		}
	})
}
