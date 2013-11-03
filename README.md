Google Analytics Dump
==========
Dumps google analytics data for all profiles, properties, and accounts into a csv using the javascript client library.

## API Keys
To apply for API access at the Google APIs Console:

- Visit the [Google APIs Console](https://code.google.com/apis/console). Log in if prompted to do so.
- Create a project for your application (if you have not already done so) by clicking Create project.
- Select Services from the menu. The list of accessible Google services appears.
- Scroll through the list until you find **Google Analytics**, click the Status switch next to the service name (so that it switches from OFF to ON.
- For some services, the Console will display a Terms of Service pane. To go ahead, check the I agree to these terms box, then click Accept.
- Scroll back to the top of the page and click API Access in the menu. 
- The API Access pane appears.
- Click Create an OAuth 2.0 client ID.
- The Create Client ID dialog appears.
- Click the Web application radio button. Type your site or hostname in the field below.
- Click more options under the radio buttons.
- The Authorized Redirect URIs and Authorized JavaScript Origins fields appear.
- Clear any text in the Authorized Redirect URIs box. (When using JavaScript, do not specify any redirects.)
- In the Authorized JavaScript Origins box, type the protocol and domain for your site.
- Make sure to enter the domain only, do not include any path value.
- If your site supports both HTTP and HTTPS, you can enter multiple values, one per line.
- Click the Create client ID button to complete the process.
- The Create client ID dialog disappears. The Authorized API Access section now displays your application's OAuth 2.0 credentials.
- Place these credentials in [js/lib/auth_utils.js](https://github.com/jfrazelle/google-analytics-dump/blob/master/js/lib/auth_utils.js)

## Build Instructions
This project uses [Grunt](http://gruntjs.com) to automate build tasks.

- Install [Node.js](http://nodejs.org)
- Install grunt-cli: `npm install -g grunt-cli`
- Install dev dependencies: `npm install`
- Run `grunt` to compile, or `grunt server` to start a live development environment.

## Configurations
To pull a different set of data change these lines in [js/main.js](https://github.com/jfrazelle/google-analytics-dump/blob/master/js/main.js) to whatever [Dimensions and Metrics](https://developers.google.com/analytics/devguides/reporting/core/dimsmets) you wish to pull.

```
gapi.client.analytics.data.ga.get({
        'ids': 'ga:' + profileId,
        'start-date': lastNDays(30),
        'end-date': lastNDays(0),
        'metrics': 'ga:pageviews, ga:visits',
        'dimensions': 'ga:date',
        //'sort': '-ga:visits,ga:source',
        //'filters': 'ga:medium==organic',
        'max-results': 500
}).execute(handleCoreReportingResults);
```
