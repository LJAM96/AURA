package jobs

import (
	downloadqueue "aura/download/queue"
	"context"
)

func StartDownloadQueueWorker(ctx context.Context) {
	downloadqueue.StartWorker(ctx)
}
