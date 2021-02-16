# Quick summary

This package provides access to master and slave MySQL servers depending on the query
type. If the query is a read-only it redirects it to one of the replicas. Otherwise it
passes the query to the master server.

## Features

- Picks up a replica server by Round Rubin algorithm.
- Performs automatic maintenance handling for replicas by 
  redirecting the query to the other replicas.

requires Go version 1.12 or greater.

## Usage

```go
func Main() {
	
    master, err := sql.Open("mysql", "username:password@tcp(127.0.0.1:3306)/test")
    if err != nil {
        panic(err.Error())
    }
    replica1, err := sql.Open("mysql", "username:password@tcp(127.0.0.2:3306)/test")
    if err != nil {
        panic(err.Error())
    }
    replica2, err := sql.Open("mysql", "username:password@tcp(127.0.0.3:3306)/test")
    if err != nil {
        panic(err.Error())
    }
    replica3, err := sql.Open("mysql", "username:password@tcp(127.0.0.4:3306)/test")
    if err != nil {
        panic(err.Error())
    }  
  
    cluster, err := sqlcluster.New(sqlcluster.Config{
        Master:       master,
    ReadReplicas: sqlcluster.Replicas{replica1, replica2, replica3},
    })
    if err != nil {
        panic(err.Error())
    }
}
```