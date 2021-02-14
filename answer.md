# Answers

## Does the library fulfill the requirements described in the background section?

- No. Replica servers can go down at any time, even in the middle of running a query.  

    __Solution: All the receivers that are calling the remote server should have
    a context with timeout. There should be a mechanism to pass the query 
    to the next available replica when time out happens.__

## Is the library easy to use?

- Easy enough to use, but it has lots of flaws that lets the user go wrong, 
  especially for a new team member.

- The package doesn't have documentation. Even simple block of `how to use` 
  code could make a huge difference in adoptability.  
  
     __adding__ 

- It doesn't encapsulate enough of the process and leaves the user an opening to 
  misuse the package quickly.  
  
     __Solution: I added `replicaPool` to handle replica servers __

## Is the code quality assured?

- looking at the fact that the package doesn't have tests is enough to not to be 
  sure about the quality.  
  
    __Solution: I'm trying to add tests. It requires creating some mocks that 
    acts like a disconnected replica. I'm figuring it out.__

- Using `panic` instead of handling and returning a proper error is a code smell.  

    __fixed__ 

- Using a slice of `interface` for replicas and asserting a member after picking 
  it up as a `*sql.DB` every single time is not a safe way to deliver a package 
  that is meant to be used in a team. It was be acceptable if the code was a 
  showcase, or possibility check. The better way to do this is to use `[]*sql.DB` 
  instead of `[]interface{}`.
      
     __Solution: I added `SQLDatabase` interface__
      
- The `NewDB` could take an interface satisfied by *sql.DB in the parameters. 
  Testing it will be possible this way.
  
    __Solution: I tried to use SQLDatabase interface as parameters but when I 
    after I added the timeout duration for the read replicas it became messy. 
    So I used a configuration struct as a parameter instead.__

- The `DB.Close` func is not returning any error.
    
    __fixed__ 
    
## Is the code readable?

- The code is readable enough, with some improvement in namings

    - `readReplicaRoundRobin` doesn't feel right. Appending `RoundRobin` in 
      the function name doesn't help readability since there is only one algorithm 
      to choose the next replica with. Maybe `pickReadReplica` or `nextReadReplica` 
      is much more readable and descriptive.  
      
      We could mention the choosing algorithm in the comments. This way if we 
      decide to change the algorithm we don't need to rename it everywhere.   
      
      __Solution: I encapsulated the iterator functionality and replica selection
      algorithm in `replicaPool` struct.__
    
    - The naming for `DB` type is misleading since it's not a database. It's more 
      like a cluster manager. `Cluster` could be more meaningful for this struct.

    __fixed. It's now called `Cluster` __ 

    - `DB.count` is an iterator that it's current naming doesn't make sense.
    
    __changed to `indexIterator.indexIterator`__ 

- Functions don't have functionality comments. Even copying a comment from another 
  package which describes the functionality is better than having no comment.
    
  __I added the comments on signatures in the `SQLDatabase` interface. 
  The **downside** is the original comments get old over time. This is something 
  that I need to get suggestion from other team members.__ 
    
## Is the library thread-safe?
  
- `db.count` faces data race problem.  

  __Solution: It could be solved by using `mutex.Lock` but I avoid using it 
  whenever I can since it makes the code a little messy and it's a little 
  bit slower compared to  `atomic` package.__

- `DB.readReplicas` doesn't have any issue since the only time it is accessed
  for writing is in `NewDB` func.

  __There's no need to do anything for now. If we decide to add dynamic replica 
  injection for in the cluster, we should add mutex lock to prevent data race 
  problem.__