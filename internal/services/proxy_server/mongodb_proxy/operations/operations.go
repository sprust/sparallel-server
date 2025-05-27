package operations

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/connections"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sync"
	"time"
)

type Operations[T OperationInterface] struct {
	name        string
	mutex       sync.Mutex
	connections *connections.Connections
	running     map[string]T
	finished    map[string]T

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
		running:     make(map[string]T),
		finished:    make(map[string]T),
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
) *objects.RunningOperation {
	opUuid := uuid.New().String()

	go func(ctx context.Context) {
		o.registerOperation(opUuid, operation)

		coll, err := o.connections.Get(connName, dbName, collName)

		if err != nil {
			operation.Error(err)
		} else {
			operation.Execute(ctx, coll)

			o.finishOperation(opUuid, operation)
		}
	}(ctx)

	return &objects.RunningOperation{Uuid: opUuid}
}

func (o *Operations[T]) Pull(uuid string) OperationInterface {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	operation, exists := o.running[uuid]

	if exists {
		return operation
	}

	operation, exists = o.finished[uuid]

	if exists {
		go o.deleteOperation(uuid, false)

		return operation
	}

	return nil
}

func (o *Operations[T]) CheckTimeouts() {
	o.timeoutsMutex.Lock()
	defer o.timeoutsMutex.Unlock()

	slog.Debug("Checking timeouts for operations: " + o.name)

	for opUuid, timeout := range o.timeouts {
		if timeout.Before(time.Now()) {
			go o.deleteOperation(opUuid, true)
		}
	}
}

func (o *Operations[T]) registerOperation(opUuid string, operation T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.running[opUuid] = operation

	go o.registerTimeout(opUuid)
}

func (o *Operations[T]) finishOperation(opUuid string, operation T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	delete(o.running, opUuid)
	o.finished[opUuid] = operation
}

func (o *Operations[T]) deleteOperation(opUuid string, isTimeout bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if isTimeout {
		slog.Warn("Deleting operation of [" + o.name + "] by timeout: " + opUuid)
	}

	delete(o.running, opUuid)
	delete(o.finished, opUuid)

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
