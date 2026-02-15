package messaging

import (
	"context"
	"log/slog"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
)

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		go p.runCreatedWorker(ctx)
	}
	for i := 0; i < p.workers; i++ {
		go p.runStatusUpdatedWorker(ctx)
	}
}

// orders.created
func (p *WorkerPool) runCreatedWorker(ctx context.Context) {
	batch := make([]core.OrderEvent, 0, p.batchSize)
	timer := time.NewTimer(p.batchTimeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-p.handler.createdQueue:
			batch = append(batch, event)
			if len(batch) >= p.batchSize {
				p.processCreatedBatch(ctx, batch)
				batch = batch[:0]
				timer.Reset(p.batchTimeout)
			}
		case <-timer.C:
			if len(batch) > 0 {
				p.processCreatedBatch(ctx, batch)
				batch = batch[:0]
			}
			timer.Reset(p.batchTimeout)
		}
	}
}

func (p *WorkerPool) processCreatedBatch(ctx context.Context, batch []core.OrderEvent) {
	for _, event := range batch {
		if err := p.repo.Create(ctx, event.ToOrder()); err != nil {
			slog.Error("[WorkerPool] repo_create_failed", "error", err, "event", event)
			if dlqErr := p.dlqProducer.PublishDLQ(ctx, event, "repo_create_failed", err); dlqErr != nil {
				slog.Error("[WorkerPool] failed to send repo_create_failed to DLQ", "error", dlqErr, "event", event)
			}
		}
	}
}

// orders.status_updated
func (p *WorkerPool) runStatusUpdatedWorker(ctx context.Context) {
	batch := make([]core.OrderEvent, 0, p.batchSize)
	timer := time.NewTimer(p.batchTimeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-p.handler.statusUpdatedQueue:
			batch = append(batch, event)
			if len(batch) >= p.batchSize {
				p.processStatusUpdatedBatch(ctx, batch)
				batch = batch[:0]
				timer.Reset(p.batchTimeout)
			}
		case <-timer.C:
			if len(batch) > 0 {
				p.processStatusUpdatedBatch(ctx, batch)
				batch = batch[:0]
			}
			timer.Reset(p.batchTimeout)
		}
	}
}

func (p *WorkerPool) processStatusUpdatedBatch(ctx context.Context, batch []core.OrderEvent) {
	for _, event := range batch {
		if err := p.repo.UpdateStatus(ctx, event.OrderID, *event.Status); err != nil {
			slog.Error("[WorkerPool] repo_update_failed", "error", err, "event", event)
			if dlqErr := p.dlqProducer.PublishDLQ(ctx, event, "repo_update_failed", err); dlqErr != nil {
				slog.Error("[WorkerPool] failed to send repo_update_failed to DLQ", "error", dlqErr, "event", event)
			}
		}
	}
}
