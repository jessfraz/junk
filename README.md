# transcribe

Convert Slack exported archive to a plain text log.

```console
$ transcribe --help
transcribe - Convert Slack exported archive to a plain text log.

 Version: v0.1.0
  -d	run in debug mode
  -i string
    	input file containing slack message archive
  -o string
    	output file for saving generated static chat log
  -u string
    	path to user json for ID conversion (optional)
  -v	print version and exit (shorthand)
  -version
    	print version and exit
```

### Example

Say you have an exported Slack private message archive in the format below
saved in a file called [`fixtures/casper.json`](fixtures/casper.json).

> **NOTE:** You can test with the ones in the [fixtures](fixtures/) directory of
> this repo.

```json
[
    {
        "type": "message",
        "user": "casper",
        "text": "Btw, i am a ghost.",
        "date": "2015-03-25T22:17:50.000Z"
    },
	{
        "type": "message",
        "user": "jessfraz",
        "text": "we have cookies!",
        "date": "2015-03-25T22:17:44.000Z"
    },
	{
        "type": "message",
        "user": "casper",
        "text": "Yo!",
        "date": "2015-03-25T22:17:38.000Z"
    }
]
```

You can use the following command to get a more readable format:

```console
$ transcribe -i fixtures/casper.json
INFO[0000] Readable chat log saved to fixtures/casper.log
```

The artifact should look like the following:

```
============================= Wednesday, March 6 2015 ============================
[15:17:38] <casper> Yo!
[15:17:44] <jessfraz> we have cookies!
[15:17:50] <casper> Btw, i am a ghost.
```

### More Interesting Example

Okay let's say you have a bit different looking export archive for a channel or
private message.

We can use the example in [`fixtures/container-cabal.json`](fixtures/container-cabal.json), which looks like the
following:

```json
[
    {
        "id": "",
        "user": "U02F7LZ5G",
        "ts": "1415233015.000338",
        "text": "See you tomorrow!",
        "type": "message"
    },
    {
        "id": "",
        "user": "U02AZ2JAM",
        "ts": "1415233026.000339",
        "text": "night",
        "type": "message"
    },
    {
        "id": "",
        "user": "U02NE6P41",
        "ts": "1415246890.000340",
        "text": "@U02F7LZ5G, @U02AZ2JAM: Awesome work.",
        "type": "message"
    }
]
```

> **NOTE:** Above we have a gross string timestamp and gross user ids.

If this is the format you got your archive in you should have also been supplied
a users.json, which looks like [`fixtures/users.json`](fixtures/users.json).
We can use that to generate a very pretty chat log output like the first
example as well.

Let's run:

```console
$ transcribe -i fixtures/container-cabal.json -u fixtures/users.json
INFO[0000] Readable chat log saved to fixtures/container-cabal.log
```

The artifact should look like the following:

```
============================= Wednesday, November 6 2014 ============================
[16:16:55] <alice> See you tomorrow!
[16:17:06] <bob> night
[20:08:10] <batman> @alice, @bob: Awesome work.
```

> **NOTE:** See how the user ids have been replaced by the more readable name ;)
