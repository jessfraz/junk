# nsq hook executer

Execute a script on an nsq message from a github hook.

```console
$ nsqexec -h
Usage of nsqexec:
  -channel="exec-hook": nsq channel
  -d=false: run in debug mode
  -exec="": path to script file to execute
  -lookupd-addr="nsqlookupd:4161": nsq lookupd address
  -topic="hooks-docker": nsq topic
  -v=false: print version and exit (shorthand)
  -version=false: print version and exit
```

Example docker run command:

```bash
$ docker run -d --restart always \
    --link nsqlookupd1:nsqlookupd \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/local/bin/docker:/usr/local/bin/docker \
    -v /tmp:/tmp \
    -e DOCKER_HOST="unix:///var/run/docker.sock" \
    --privileged \
    -v /path/to/script.sh:/path/to/script.sh
    --name nsqexec \
    jess/nsqexec -d -exec="/path/to/script.sh" \
    -topic hooks-docker -channel hook \
    -lookupd-addr nsqlookupd:4161
```
