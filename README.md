## Feed Script

Get all Twitter statuses and Instagram photos for a specific user and update with latest over time, avoid being rate limited with cron.

Saves all to a file that is a JSON list of objects, in order of most recent first.

### Setup Credentials

Add your credentials from the following auths into the **[top of main.py](main.py#L14-26)**.

#### Create a Twitter application
1. Sign in with [Twitter Developer](https://dev.twitter.com/)
2. Hover over your name in the top right corner then click "My Applications"
3. Create a New Application. Enter a name (this is for your reference), a description (again for your reference), and your site's URL. The callback URL is a moot point for the use of the application so it can be left blank.
4. Create my Access Token (this is a button, click it)

#### Create an Instagram application
1. Sign in with [Instagram Developer](http://instagram.com/developer/)
2. Click "Register Your Application" (this is a button, click it).
3. Create a New Application. Enter a name (this is for your reference), a description (again for your reference), your site's URL, and the url for the directory ```instagram``` where your app will live for the redirect uri.

### Configure

Change the user you wish to query for in **[main.py](main.py#L151-152)**

### Install on server

```bash
# make sure you are running python 2.7
$ python --version
# Python 2.7.(Any)

# install pip (if you don't already have it)
$ sudo apt-get install python-pip

# install the dependencies from requirements.txt
$ pip install -r requirements.txt

# try the script
$ ./main.py
```

### Results format

The results are saved to a file named `feed.json`. It is formatted like the following:

```javascript
[
  {
    "text": "Yes. #newyorker http://t.co/6aSWzJvA1g",
    "created_time": "2014-05-09 11:48:27",
    "retweet_count": 0,
    "type": "tweet",
    "id": "464733635904806912",
    "favorite_count": 0
  },
  {
    "filter": "Valencia",
    "caption": "There are perks to being this close to the italian border",
    "like_count": 7,
    "link": "http://instagram.com/p/nnfS3PSr0w/",
    "location": {
      "point": {
        "latitude": 43.6972298,
        "longitude": 7.27630694
      },
      "id": 2041542,
      "name": "Fenocchio Glacier"
    },
    "created_time": "2014-05-05 13:35:14",
    "images": {
      "low_resolution": {
        "url": "http://origincache-ash.fbcdn.net/1389592_231849820343549_2086542155_a.jpg",
        "width": 306,
        "height": 306
      },
      "thumbnail": {
        "url": "http://origincache-ash.fbcdn.net/1389592_231849820343549_2086542155_s.jpg",
        "width": 150,
        "height": 150
      },
      "standard_resolution": {
        "url": "http://origincache-ash.fbcdn.net/1389592_231849820343549_2086542155_n.jpg",
        "width": 640,
        "height": 640
      }
    },
    "type": "insta",
    "id": "713676701666295088_4714782"
  },
  ...
]
```

### Setup Cron Job

The cron job will run every 5 minutes.

```bash
$ crontab -e

# add the following line to the file
*/5 * * * * /path/to/script/main.py
```

### Accessing through javascript

You can ajax load the file into your app, and iterate through the items.

#### Vanilla Javascript

```javascript
request = new XMLHttpRequest();
request.open('GET', 'route/to/feed.json', true);

request.onload = function() {
  if (request.status >= 200 && request.status < 400){
    feed = JSON.parse(request.responseText);
    feed.forEach(function(item){
      // do something with the item
      console.log(item);
    });
  } else {
    // handle error
  }
};

request.onerror = function() {
  // handle error
};

request.send();
```

#### Vanilla Javascript

```javascript
$.ajax({
  type: "GET",
  url: "route/to/feed.json",
  dataType: "json"
}).done(function(feed) {
  feed.forEach(function(item){
    // do something with the item
    console.log(item);
  });
}).fail(function(jqXHR, textStatus) {
  // handle failure
  console.log("Request failed: " + textStatus);
});;
```