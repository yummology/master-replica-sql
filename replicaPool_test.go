package sqlcluster

import (
	`database/sql`
	`errors`
	`log`
	`reflect`
	`testing`
	`time`
)

// 3 replicas, none of them are down
func TestReplicaPoolNextIndexNoFails(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 2, 0, 1, 2, 0}) {
		log.Println("no replica should be skipped")
		t.FailNow()
	}
}

// 4 replica, second one is down
func TestReplicaPoolNextIndex1(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 1)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 2, 3, 0, 2, 3, 0}) {
		log.Println("replica #1 should be skipped")
		t.FailNow()
	}
}

// 4 replica, 3rd one is down
func TestReplicaPoolNextIndex2(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 2)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 3, 0, 1, 3, 0}) {
		log.Println("replica #2 should be skipped")
		t.FailNow()
	}
}

var queryFailed = errors.New("query failed")

func TestReplicaPoolReplica0UnderMaintenance(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.testMode = true
	err := pool.RunOnNextReplica(func(i int, replica SQLDatabase) error {
		if i == 0 {
			return sql.ErrConnDone
		}
		return nil
	})
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		log.Println("query must be passed to the next replica, and it should not return error")
		t.FailNow()
	}
	if !pool.isInMaintenanceMode {
		log.Println("cluster should go maintenance mode")
		t.FailNow()
	}
	if pool.underMaintenanceReplica != 0 {
		log.Println("replica #0 should be flagged as under maintenance")
		t.FailNow()
	}
}

func TestReplicaPoolReplica0ReturnsError(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.testMode = true
	err := pool.RunOnNextReplica(func(i int, replica SQLDatabase) error {
		if i == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
	time.Sleep(time.Millisecond * 50)
	if err == nil {
		log.Println("replica 0 must return sql.ErrNoRows")
		t.FailNow()
	}
	if pool.isInMaintenanceMode {
		log.Println("cluster SHOULD NOT go maintenance mode")
		t.FailNow()
	}
	if pool.underMaintenanceReplica == 0 {
		log.Println("replica #0 SHOULD NOT be flagged as under maintenance")
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
