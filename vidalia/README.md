# vidalia

Downloaded a list of TOR exit nodes and classify them via the
[Cloudflare Threat Score API](https://www.cloudflare.com/docs/client-api.html#s3.6).

```
$ vidalia --help
       _     _       _ _
__   _(_) __| | __ _| (_) __ _
\ \ / / |/ _` |/ _` | | |/ _` |
 \ V /| | (_| | (_| | | | (_| |
  \_/ |_|\__,_|\__,_|_|_|\__,_|

 Downloaded a list of TOR exit nodes and classify them via
 the Cloudflare Threat Score API
 Version: v0.1.0

  -apikey string
        Cloudflare API Key
  -c int
        Number of concurrent lookups to run (default 10)
  -d    run in debug mode
  -email string
        Cloudflare Email
  -esuri string
        Connection string for elastic search cluster (ie: tcp://localhost:9300)
  -v    print version and exit (shorthand)
  -version
        print version and exit
```
