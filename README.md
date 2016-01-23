# hulk

Ultimate job/build runner -OR- bash execution as a service.


```console
# start the server
$ hulk -d server 

# start the server with a specific socket
$ hulk -d -a /hulk.sock server 

# list jobs
$ hulk ls

# start a job
$ hulk start --name jess_is_cool --cmd "echo jess is cool"
1 <--- this is the job ID

# get a jobs logs
$ hulk logs --id 1
jess is cool

# list jobs again
$ hulk ls
ID                  NAME                CMD                 STATUS              ARTIFACTS
1                   jess_is_cool        echo jess is cool   completed

# get the job state
$ hulk state --id 12
running

# delete a job
$ hulk rm --id 12
```

**TODO:**
- way to filter in list
- artifact moving on success
- designate whether to email on error or on success
- scheduler
- build steps
