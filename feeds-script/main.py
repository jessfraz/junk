#!/usr/bin/env python

import moarjson as json
from datetime import datetime
from time import mktime
json.register(datetime, str)

from os import path
from operator import itemgetter
import tweepy as t
from instagram.client import InstagramAPI

file = 'feed.json'
creds = {
    'twitter': {
        'api_key': 'YOUR_API_KEY',
        'api_secret': 'YOUR_SECRET',
        'access_token': 'YOUR_TOKEN',
        'access_token_secret': 'YOUR_TOKEN_SECRET'
    },
    'instagram': {
        'client_id': 'YOUR_CLIENT_ID',
        'client_secret': 'YOUR_CLIENT_SECRET',
        'access_token': 'YOUR_ACCESS_TOKEN'
    }
}


# initialize feed file with an empty object
def init_file():
    with open(file, mode='w+') as f:
        json.dump([], f, indent=2)


# get users tweets
def getTweets(username, last):
    # authenticate
    twitter_creds = creds['twitter']
    auth = t.OAuthHandler(twitter_creds['api_key'],
                          twitter_creds['api_secret'])
    auth.set_access_token(twitter_creds['access_token'],
                          twitter_creds['access_token_secret'])
    api = t.API(auth)

    tweets = []
    try:
        for s in t.Cursor(api.user_timeline, screen_name=username,
                          since_id=last).items():
            # prepend to feed
            s = s._json
            tweets.append({
                'type': 'tweet',
                'id': s['id_str'],
                'created_time': datetime.strptime(s['created_at'],
                                                  "%a %b %d %X +0000 %Y"),
                'favorite_count': s['favorite_count'],
                'retweet_count': s['retweet_count'],
                'text': s['text']
            })
    except Exception, e:
        print e
    return tweets


# get users instagrams
def getInstagrams(userid, last, init):
    insta_creds = creds['instagram']
    api = InstagramAPI(client_id=insta_creds['client_id'],
                       client_secret=insta_creds['client_secret'])

    photos = []
    try:
        api = InstagramAPI(access_token=insta_creds['access_token'])
        media_feed, next = api.user_recent_media(user_id=userid)
        for m in media_feed:
            photos = instaAppend(m, photos, last, init)
        counter = 1
        while next and counter < 100:
            media_feed, next = api.user_recent_media(user_id=userid,
                                                     with_next_url=next)
            for m in media_feed:
                photos = instaAppend(m, photos, last, init)
            counter += 1
    except Exception, e:
        print e
    return photos


# clean insta
def instaAppend(m, photos, last, init):
    m = m.__dict__
    print "is new"
    print init
    print "is greater"
    print mktime(m['created_time'].timetuple()) > last
    if mktime(m['created_time'].timetuple()) > last or init is True:
        if "location" in m:
            loc = m['location'].__dict__
            loc['point'] = loc['point'].__dict__
        else:
            loc = {}
        if m['caption'] is not None:
            m['caption'] = m['caption'].__dict__['text']
        photos.append({
            'type': 'insta',
            'id': m['id'],
            'created_time': m['created_time'],
            'caption': m['caption'],
            'filter': m['filter'],
            'images': {
                'low_resolution': m['images']['low_resolution'].__dict__,
                'standard_resolution':
                m['images']['standard_resolution'].__dict__,
                'thumbnail': m['images']['thumbnail'].__dict__
            },
            'like_count': m['like_count'],
            'link': m['link'],
            'location': loc,
        })
    return photos

if __name__ == '__main__':
    feed = []
    last_tweet = None
    last_insta = 0
    init = False

    # check if feeds file exists, if not create it
    # otherwise read the existing file
    #   parse for most recent tweet saved
    #   parse for most recent insta saved
    if (path.isfile(file) is False):
        init_file()
        init = True
    else:
        with open(file, mode='r') as feedjson:
            feed = json.load(feedjson)
        tweet_found = False
        insta_found = False
        for v in feed:
            if v['type'] == 'tweet' and tweet_found is False:
                last_tweet = v['id']
                tweet_found = True
            if v['type'] == 'insta' and insta_found is False:
                last_insta = mktime(datetime.strptime(v['created_time'],
                                    "%Y-%m-%d %X").timetuple())
                insta_found = True
            if tweet_found is True and insta_found is True:
                break

    tweets = getTweets('frazelledazzell', last_tweet)
    insta = getInstagrams('4714782', last_insta, init)

    rsorted = sorted((tweets+insta), key=itemgetter('created_time'),
                     reverse=True)

    feed = rsorted + feed

    with open(file, mode='w') as feedjson:
        json.dump(feed, feedjson, indent=2)

    print "added " + str(len(tweets)) + " tweets and " \
          + str(len(insta)) + " photos"
