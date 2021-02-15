# Answers

## Does the library fulfill the requirements described in the background section?

- No, it doesn't. Replica servers can go down at any time, even in the middle of
  executing a query. So the package should pass the query to the next replica if
  this happens.  

    __Solution: I added a mechanism that checks tries passing the query to the
    next replica when `sql.ErrConnDone` happens.__

## Is the library easy to use?

- Easy enough to use, but with a configuration struct there will be no point to
  go wrong with setting it up, especially for a new team member.
  
     __Solution: configuration struct added__
  
- The package doesn't have documentation. Even simple block of `how to use` 
  code could make a huge difference in adoptability. Most of the time developers
  don't even bother trying a package that doesn't have readme file.
  
     __Solution: README.md file added__

## Is the code quality assured?

- Using `panic` instead of handling and returning a proper error is a code smell.  

    __fixed__ 

- Using a slice of `interface` for replicas and asserting a member after picking 
  it up as a `*sql.DB` every single time is not a safe way to deliver a package 
  that is meant to be used in a team. It was be acceptable if the code was a 
  showcase, or possibility check. The better way to do this is to use `[]*sql.DB` 
  instead of `[]interface{}`.
      
     __Solution: The `NewDB` could take an interface satisfied by `*sql.DB` in 
     the parameters. I added `SQLDatabase` interface so testing it will be 
     possible this way.__

- looking at the fact that the package doesn't have tests is enough to not to be 
  sure about the quality.  
  
    __Solution: I added tests for the replicaPool functionality. But, to test the
    cluster itself it requires using mockery using `github.com/golang/mock`.
    By adding and using the `SQLDatabase` interface, it's possible to generate a
    mock struct for it by using `mockgen`, but I haven't tried this package 
    for mocking yet.__

- The `DB.Close` func is not returning any error.
    
    __fixed__ 
    
## Is the code readable?

The code is readable enough, with some improvement in namings

- The naming for `DB` type is misleading since it's not a database. It's more 
  like a cluster manager. `Cluster` could be more meaningful for this struct.

     __fixed. It's now called `Cluster` __ 

- It doesn't encapsulate enough of the process and leaves the user an opening to 
  misuse the package quickly.  

     __Solution: I added `replicaPool` to handle proxying replica servers, and 
     detecting if they are in maintenance state__
 
- `readReplicaRoundRobin` doesn't feel a right naming. Appending `RoundRobin` in 
  the function name doesn't help readability since there is only one algorithm 
  to choose the next replica with. We could mention the selection algorithm in 
  the comments. This way if we decide to change the algorithm we don't need 
  to rename function calls everywhere.   
  
     __Solution: I encapsulated the iterator functionality and replica selection
     algorithm in `replicaPool` struct.__

- `DB.count` is an iterator that it's current naming doesn't make sense.

     __changed to `replicaPool.iterator`__

- Functions don't have functionality comments. Even copying a comment from the
  `sql` package which describes the functionality is better than having no comment.
    
     __I added the comments.
     The **downside** is the original comments get old over time. This is something 
     that I need to get suggestion from other team members.__ 
    
## Is the library thread-safe?
  
- `db.count` faces data race problem.  

  __Solution: It could be solved by using `mutex.Lock` but I avoid using it 
  whenever I can since it makes the code a little messy, and it's a little slower 
  compared to `atomic` package.__

- `DB.readReplicas` doesn't have any issue since the only time it is accessed
  for writing is on creation time.

  __There's no need to do anything for now. If we decide to add dynamic replica 
  injection for in the cluster, we should add mutex lock to prevent data race.__