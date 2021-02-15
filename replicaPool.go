package sqlcluster

import (
	`context`
	`database/sql`
	`errors`
	`sync`
	`sync/atomic`
	`time`
)

const iteratorStep uint64 = 1

type replicaPool struct {
	replicas                []SQLDatabase
	iterator                uint64
	replicasCount           uint64
	underMaintenanceReplica int
	timeout                 time.Duration
	isInMaintenanceMode     bool
	mutex                   sync.Mutex
	testMode                bool
}

// newReplicaPool is simple `replicaPool` factory
func newReplicaPool(timeout time.Duration, replicas ...SQLDatabase) (*replicaPool, error) {
	if len(replicas) < 2 {
		return nil, errors.New("minimum number of replicas servers should be 2")
	}
	return &replicaPool{
		replicas:      replicas,
		iterator:      uint64(len(replicas) - 1),
		timeout:       timeout,
		replicasCount: uint64(len(replicas)),
	}, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (pool *replicaPool) Ping() (err error) {
	for index := range pool.replicas {
		if Err := pool.replicas[index].Ping(); Err != nil {
			err = Err
			go pool.maintenanceHandler(index)
		}
	}
	return err
}

// PingContext verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (pool *replicaPool) PingContext(ctx context.Context) (err error) {
	for index := range pool.replicas {
		if Err := pool.replicas[index].PingContext(ctx); Err != nil {
			err = Err
			go pool.maintenanceHandler(index)
			break
		}
	}
	return err
}

// nextIndex returns next available read replica index.
// If the replica is under maintenance it skips to the next index.
// The algorithm for choosing the next replica is Round Rubin.
func (pool *replicaPool) nextIndex() int {
	index := int(atomic.AddUint64(&pool.iterator, iteratorStep) % pool.replicasCount)
	for pool.isInMaintenanceMode {
		if index != pool.underMaintenanceReplica {
			return index
		}
		index = int(atomic.AddUint64(&pool.iterator, iteratorStep) % pool.replicasCount)
	}
	return index
}

// RunOnNextReplica provides the next selected replica and a context as a parameter to a function
func (pool *replicaPool) RunOnNextReplica(
	mainContext context.Context,
	fn func(ctx context.Context, replicaIndex int, replica SQLDatabase) error,
) error {

	for true {

		replicaCTX, cancel := context.WithTimeout(context.Background(), pool.timeout)
		index := pool.nextIndex()
		chErr := make(chan error)
		go func() {
			chErr <- fn(mainContext, index, pool.replicas[index])
		}()

		select {
		// request has been canceled, by any possible reasons
		case <-mainContext.Done():
			cancel()
			return mainContext.Err()

		// replica timeout happened
		case <-replicaCTX.Done():
			go pool.maintenanceHandler(index)

		// execution finished
		case err := <-chErr:
			if ! errors.Is(err , sql.ErrConnDone) {
				return err
			}
			go pool.maintenanceHandler(index)
		}
	}

	return nil
}

// maintenanceHandler handles the replica server maintenance flag.
// and runs a watcher over it. Also makes ensure only one instance of watcher executes.
func (pool *replicaPool) maintenanceHandler(index int) {
	// mutex is used here to make sure only one instance of
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	if pool.isInMaintenanceMode {
		return
	}
	pool.setMaintenanceFlag(true, index)
	if ! pool.testMode {
		go pool.watchReplica(index)
	}
}

// watchReplica pings the replica under maintenance every second and as
// soon as getting response removes the maintenance flag over it
func (pool *replicaPool) watchReplica(index int) {
	ticker := time.NewTicker(time.Second)
	return
	for {
		<-ticker.C
		if err := pool.replicas[index].Ping(); err == nil {
			pool.setMaintenanceFlag(false, -1)
			ticker.Stop()
			return
		}
	}
}

// setMaintenanceFlag sets replica's under-maintenance flag
func (pool *replicaPool) setMaintenanceFlag(underMaintenance bool, index int) {
	pool.underMaintenanceReplica = index
	pool.isInMaintenanceMode = underMaintenance
}

// Walk runs a func over all replica servers
func (pool *replicaPool) Walk(fn func(replica SQLDatabase)) {
	for index := range pool.replicas {
		fn(pool.replicas[index])
	}
}
