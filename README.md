# callmemaybe

Stupid demo to show inital seccomp support in docker.

```console
$ make docker

$ docker run --rm -it --security-opt seccomp:$(pwd)/callmemaybe.json jess/callmemaybe
operation not permitted
docker: Error response from daemon: Cannot start container ba7ef732c312ef53156274b22da34a2a37d65ed7606cf33b4af7ffa542c0ee37: [8] System error: operation not permitted.
```
