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
	conf := sqlcluster.Config {

    }
    sqlcluster.New()


}
```

### What is this repository for? ###

* Quick summary
* Version
* [Learn Markdown](https://bitbucket.org/tutorials/markdowndemo)

### How do I get set up? ###

* Summary of set up
* Configuration
* Dependencies
* Database configuration
* How to run tests
* Deployment instructions

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### Who do I talk to? ###

* Repo owner or admin
* Other community or team contact