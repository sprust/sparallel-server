package operations

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/connections"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sync"
	"time"
)

type Operations[T OperationInterface] struct {
	name        string
	mutex       sync.Mutex
	connections *connections.Connections
	items       map[string]T

	timeoutsMutex sync.Mutex
	timeouts      map[string]time.Time
	ticker        *time.Ticker
}

func NewOperations[T OperationInterface](
	ctx context.Context,
	name string,
	connections *connections.Connections,
) *Operations[T] {
	o := &Operations[T]{
		name:        name,
		connections: connections,
		items:       make(map[string]T),
		timeouts:    make(map[string]time.Time),
		ticker:      time.NewTicker(10 * time.Second),
	}

	go func(ctx context.Context, o *Operations[T]) {
		defer o.ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-o.ticker.C:
				o.CheckTimeouts()
			}
		}
	}(ctx, o)

	return o
}

func (o *Operations[T]) Add(
	ctx context.Context,
	connName string,
	dbName string,
	collName string,
	operation T,
) *mongodb_proxy_objects.RunningOperation {
	opUuid := uuid.New().String()

	go func(ctx context.Context) {
		o.registerOperation(opUuid, operation)

		coll, err := o.connections.Get(connName, dbName, collName)

		if err != nil {
			slog.Error(o.name + "[" + opUuid + "]: failed to get collection: " + err.Error())

			operation.Error(err)
		} else {
			operation.Execute(ctx, coll)

			slog.Debug(o.name + "[" + opUuid + "]: executed")
		}
	}(ctx)

	return &mongodb_proxy_objects.RunningOperation{Uuid: opUuid}
}

func (o *Operations[T]) Pull(ctx context.Context, opUuid string) (OperationInterface, string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	operation, exists := o.items[opUuid]

	if !exists || !operation.IsFinished() {
		slog.Debug(o.name + "[" + opUuid + "]: pulled nil")

		return nil, ""
	}

	nextOpUuid := ""

	if operation.HasNext() {
		nextOpUuid = uuid.New().String()

		go func() {
			slog.Debug(o.name + "[" + opUuid + "]: cloning as [" + nextOpUuid + "]")

			cloned := operation.Clone(ctx).(T)

			o.registerOperation(nextOpUuid, cloned)

			slog.Debug(o.name + "[" + nextOpUuid + "]: executed [cloned from [" + opUuid + "]")
		}()
	}

	go o.deleteOperation(opUuid, false)

	slog.Debug(o.name + "[" + opUuid + "]: pulled")

	return operation, nextOpUuid
}

func (o *Operations[T]) CheckTimeouts() {
	o.timeoutsMutex.Lock()
	defer o.timeoutsMutex.Unlock()

	slog.Debug(o.name + ": checking timeouts for operations")

	for opUuid, timeout := range o.timeouts {
		if timeout.Before(time.Now()) {
			go o.deleteOperation(opUuid, true)
		}
	}
}

func (o *Operations[T]) Stats() mongodb_proxy_objects.OperationsStats {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return mongodb_proxy_objects.OperationsStats{
		Items:    len(o.items),
		Timeouts: len(o.timeouts),
	}
}

func (o *Operations[T]) registerOperation(opUuid string, operation T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.items[opUuid] = operation

	go o.registerTimeout(opUuid)
}

func (o *Operations[T]) deleteOperation(opUuid string, isTimeout bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if isTimeout {
		slog.Warn("Deleting operation [" + opUuid + "] of [" + o.name + "] by timeout")
	}

	delete(o.items, opUuid)

	go o.deleteTimeout(opUuid)
}

func (o *Operations[T]) registerTimeout(opUuid string) {
	o.timeoutsMutex.Lock()
	defer o.timeoutsMutex.Unlock()

	o.timeouts[opUuid] = time.Now().Add(5 * time.Minute)
}

func (o *Operations[T]) deleteTimeout(opUuid string) {
	o.timeoutsMutex.Lock()
	defer o.timeoutsMutex.Unlock()

	delete(o.timeouts, opUuid)
}
