package sqlcluster

import (
	`database/sql`
	`log`
	`reflect`
	`testing`
	`time`
)

// 3 replicas, none of them are down
func TestReplicaPoolNextIndexNoFails(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 2, 0, 1, 2, 0}) {
		log.Println("no replica MUST be skipped")
		t.FailNow()
	}
}

// 4 replica, second one is down
func TestReplicaPoolNextIndex1(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 1)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 2, 3, 0, 2, 3, 0}) {
		log.Println("replicaPool MUST skip replica #1")
		t.FailNow()
	}
}

// 4 replica, 3rd one is down
func TestReplicaPoolNextIndex2(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.setMaintenanceFlag(true, 2)
	if !reflect.DeepEqual(getIndexes(7, pool), []int{0, 1, 3, 0, 1, 3, 0}) {
		log.Println("replicaPool MUST skip replica #2")
		t.FailNow()
	}
}

func TestReplicaPoolReplica0UnderMaintenance(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.testMode = true
	err := pool.RunOnNextReplica(func(i int, _ SQLDatabase) error {
		if i == 0 {
			// return sql.ErrConnDone from replica #0
			// fake it like it's disconnected
			return sql.ErrConnDone
		}
		return nil
	})
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		log.Println("replicaPool MUST pass the query to the next replica, and it SHOULD NOT return error")
		log.Println("but it returns:", err)
		t.FailNow()
	}
	if !pool.isInMaintenanceMode {
		log.Println("replicaPool MUST go maintenance mode")
		t.FailNow()
	}
	if pool.underMaintenanceReplica != 0 {
		log.Println("replicaPool MUST flag replica #0 under maintenance")
		t.FailNow()
	}
}

func TestReplicaPoolReplica0ReturnsError(t *testing.T) {
	pool, _ := newReplicaPool(nil, nil, nil, nil)
	pool.testMode = true
	err := pool.RunOnNextReplica(func(i int, _ SQLDatabase) error {
		if i == 0 {
			// return sql.ErrNoRows from replica #0
			return sql.ErrNoRows
		}
		return nil
	})
	time.Sleep(time.Millisecond * 50)
	if err == nil {
		log.Println("replicaPool MUST return sql.ErrNoRows from replica #0")
		log.Println("but it returns ", err)
		t.FailNow()
	}
	if pool.isInMaintenanceMode {
		log.Println("replicaPool SHOULD NOT go maintenance mode")
		t.FailNow()
	}
	if pool.underMaintenanceReplica == 0 {
		log.Println("replicaPool MUST NOT flag replica #0 under maintenance")
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
