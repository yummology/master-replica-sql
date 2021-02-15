package sqlcluster

import (
	`context`
	`errors`
	`reflect`
	`testing`
	`time`
)

// the idea to test this package is to mock SQLDatabase servers using
// "github.com/golang/mock/gomock" package
//
// we can use 2 replicas, the first one should through timeout error
// and the second one should response normally.
// since on the creation of the replic

// 3 replicas, none of them are down
func TestReplicaPool_nextIndex(t *testing.T) {
	pool, _ := newReplicaPool(time.Second, nil, nil, nil)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 2, 0, 1, 2, 0}) {
		t.FailNow()
	}
}

// 4 replica, second one is down
func TestReplicaPool_nextIndex1(t *testing.T) {
	pool, _ := newReplicaPool(time.Second, nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 1)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 2, 3, 0, 2, 3, 0}) {
		t.FailNow()
	}
}

// 4 replica, 3rd one is down
func TestReplicaPool_nextIndex2(t *testing.T) {
	pool, _ := newReplicaPool(time.Second, nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 2)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 3, 0, 1, 3, 0}) {
		t.FailNow()
	}
}

var queryFailed = errors.New("query failed")

// the first replica should take
func TestReplicaPool_RunOnNextReplica(t *testing.T) {
	pool, _ := newReplicaPool(time.Millisecond*200, nil, nil, nil, nil)
	pool.testMode = true
	err := pool.RunOnNextReplica(
		context.Background(),
		waitInNthReplicaFor(0, time.Millisecond*300),
	)
	if err != nil {
		t.FailNow()
	}
	if !pool.isInMaintenanceMode {
		t.FailNow()
	}
	if pool.underMaintenanceReplica != 0 {
		t.FailNow()
	}
}

// getIndexes gets N indexes from pool and returns in a slice
func getIndexes(n int, pool *replicaPool) (indexes []int) {
	for i := 0; i < n; i++ {
		indexes = append(indexes, pool.nextIndex())
	}
	return
}

func waitInNthReplicaFor(index int, d time.Duration) func(ctx context.Context, i int, replica SQLDatabase) error {
	return func(ctx context.Context, i int, replica SQLDatabase) error {
		chErr := make(chan error)
		go func() {
			if i == index {
				time.Sleep(d)
				chErr <- queryFailed
			}
			chErr <- nil
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-chErr:
			return err
		}
		return nil
	}
}
