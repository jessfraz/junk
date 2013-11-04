/*
analytics-dump - v - 2013-11-03
Dumps analytics from google into a csv
Lovingly coded by Jess Frazelle  - http://frazelledazzell.com/ 
*/
/* Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * @fileoverview Utility for handling authorization and updating the UI
 * accordingy.
 * @author api.nickm@gmail.com (Nick Mihailovski)
 */



/**
 * Authorization information. This should be obtained through the Google APIs
 * developers console. https://code.google.com/apis/console/
 * Also there is more information about how to get these in the authorization
 * section in the Google JavaScript Client Library.
 * https://code.google.com/p/google-api-javascript-client/wiki/Authentication
 */
var clientId = '1008856934158.apps.googleusercontent.com';
var apiKey = 'AIzaSyC9IVFnVPB5_1X7uUXq90UvJUYeJSHmNiU';
var scopes = 'https://www.googleapis.com/auth/analytics.readonly';


/**
 * Callback executed once the Google APIs Javascript client library has loaded.
 * The function name is specified in the onload query parameter of URL to load
 * this library. After 1 millisecond, checkAuth is called.
 */
function handleClientLoad() {
  gapi.client.setApiKey(apiKey);
  window.setTimeout(checkAuth, 1);
}


/**
 * Uses the OAuth2.0 clientId to query the Google Accounts service
 * to see if the user has authorized. Once complete, handleAuthResults is
 * called.
 */
function checkAuth() {
  gapi.auth.authorize({
    client_id: clientId, scope: scopes, immediate: true}, handleAuthResult);
}


/**
 * Handler that is called once the script has checked to see if the user has
 * authorized access to their Google Analytics data. If the user has authorized
 * access, the analytics api library is loaded and the handleAuthorized
 * function is executed. If the user has not authorized access to their data,
 * the handleUnauthorized function is executed.
 * @param {Object} authResult The result object returned form the authorization
 *     service that determine whether the user has currently authorized access
 *     to their data. If it exists, the user has authorized access.
 */
function handleAuthResult(authResult) {
  if (authResult) {
    gapi.client.load('analytics', 'v3', handleAuthorized);
  } else {
    handleUnauthorized();
  }
}


/**
 * Updates the UI once the user has authorized this script to access their
 * data. This changes the visibiilty on some buttons and adds the
 * makeApiCall click handler to the run-demo-button.
 */
function handleAuthorized() {
  var authorizeButton = document.getElementById('authorize-button');
  var runDemoButton = document.getElementById('make-api-call-button');

  authorizeButton.style.display = 'none';
  runDemoButton.style.display = 'block';
  runDemoButton.onclick = makeApiCall;
  outputToPage('Click the Get Data button to begin.');
}


/**
 * Updates the UI if a user has not yet authorized this script to access
 * their Google Analytics data. This function changes the visibility of
 * some elements on the screen. It also adds the handleAuthClick
 * click handler to the authorize-button.
 */
function handleUnauthorized() {
  var authorizeButton = document.getElementById('authorize-button');
  var runDemoButton = document.getElementById('make-api-call-button');

  runDemoButton.style.display = 'none';
  authorizeButton.style.display = 'block';
  authorizeButton.onclick = handleAuthClick;
  outputToPage('Please authorize this script to access Google Analytics.');
}


/**
 * Handler for clicks on the authorization button. This uses the OAuth2.0
 * clientId to query the Google Accounts service to see if the user has
 * authorized. Once complete, handleAuthResults is called.
 * @param {Object} event The onclick event.
 */
function handleAuthClick(event) {
  gapi.auth.authorize({
    client_id: clientId, scope: scopes, immediate: false}, handleAuthResult);
  return false;
}

function makeApiCall() {
    outputToPage('Querying Accounts.');
    gapi.client.analytics.management.accounts.list().execute(handleAccounts);
}


function handleAccounts(response) {
    if (!response.code) {
        if (response && response.items && response.items.length) {
            var max = response.items.length;
            for(var i=0; i < max; i++){
                var account_id = response.items[i].id;
                queryWebproperties(account_id, i, max);
            }
        } else {
            updatePage('No accounts found for this user.')
        }
    } else {
        updatePage('There was an error querying accounts: ' + response.message);
    }
}



var accountsDone = false;
function queryWebproperties(accountId, i, max) {
    setTimeout(function(){
        updatePage('Querying Webproperties '+(i+1)+' of '+max+'.');
        gapi.client.analytics.management.webproperties.list({
            'accountId': accountId
        }).execute(handleWebproperties);
        if (i>=(max-1)){
            accountsDone = true;
            //console.log('accounts done', i, max);
        } else {
            accountsDone = false;
        }
    }, 3000*i);
}


function handleWebproperties(response) {
    if (!response.code) {
        if (response && response.items && response.items.length) {
            var max=response.items.length;
            for(var i=0; i< max; i++){
                var account_id = response.items[i].accountId;
                var web_property_id = response.items[i].id;
                queryProfiles(account_id, web_property_id, i, max);
            }
        } else {
            updatePage('No webproperties found for this user.')
        }
    } else {
        updatePage('There was an error querying webproperties: ' + response.message);
    }
}


var propertiesDone = false;
function queryProfiles(accountId, webpropertyId, i, max) {
    setTimeout(function(){
        updatePage('Querying Profiles '+(i+1)+' of '+max+'.');
        gapi.client.analytics.management.profiles.list({
            'accountId': accountId,
            'webPropertyId': webpropertyId
        }).execute(handleProfiles);
        if (accountsDone && i>=(max-1)){
            propertiesDone = true;
            //console.log('properties done', i, max);
        } else {
            propertiesDone = false;
        }
    }, 3000*i);
}



function handleProfiles(response) {
    if (!response.code) {
        if (response && response.items && response.items.length) {
            var max = response.items.length;
            for(var i=0; i< max; i++){
                var profile_id = response.items[i].id;
                queryCoreReportingApi(profile_id, i, max);
            }

        } else {
            updatePage('No profiles found for this user.')
        }
    } else {
        updatePage('There was an error querying profiles: ' + response.message);
    }
}


var profilesDone = false;
function queryCoreReportingApi(profileId, i, max) {
    setTimeout(function(){
        updatePage('Querying Core Reporting API.');
        gapi.client.analytics.data.ga.get({
            'ids': 'ga:' + profileId,
            'start-date': lastNDays(30),
            'end-date': lastNDays(0),
            'metrics': 'ga:pageTitle, ga:pageviews, ga:visits',
            'dimensions': 'ga:date',
            //'sort': '-ga:visits,ga:source',
            //'filters': 'ga:medium==organic',
            'max-results': 500
        }).execute(handleCoreReportingResults);
        if (accountsDone && propertiesDone && i>=(max-1)){
            profilesDone = true;
            //console.log('profiles done', i, max);
        } else {
            profilesDone = false;
        }
    }, 3000*i);
}


var headersDone = false;
var output = [];
var csv_array = [];
function handleCoreReportingResults(response) {
    if (!response.code) {
        if (response.rows && response.rows.length) {
            resultsToPage('Adding Results for Profile Name: '+response.profileInfo.profileName+'...<br>');

            // Put headers in table.
            if (!headersDone){
                var header_array = [];
                output.push('<table class="table table-condensed table-striped"><thead>');
                output.push('<tr>');
                output.push('<th>Profile Name</th>');
                header_array.push('Profile Name');
                for (var i = 0, header; header = response.columnHeaders[i]; ++i) {
                    output.push('<th>', header.name, '</th>');
                    header_array.push(header.name);
                }
                headersDone = true;
                output.push('</tr></thead><tbody>');
                csv_array.push(header_array);
            }            

            // Put cells in table.

            for (var i = 0, row; row = response.rows[i]; ++i) {
                var row_array = row;
                output.push('<tr><td>'+response.profileInfo.profileName+'</td><td>', row.join('</td><td>'), '</td></tr>');
                row_array.unshift(response.profileInfo.profileName);
                csv_array.push(row_array);
            }
            resultsToPage(output.join(''));
        } else {
            outputToPage('No results found.');
        }
    } else {
        updatePage('There was an error querying core reporting API: ' + response.message);
    }
}


function outputToPage(output) {
    document.getElementById('output').innerHTML = output;
}

function resultsToPage(output) {
    if (accountsDone && profilesDone && propertiesDone){
        //console.log('i think im done');
        document.getElementById('output').innerHTML = 'Creating CSV...';
        createCSV(csv_array);
    } else  {
        document.getElementById('output').innerHTML = 'Querying Next...';
    }
    document.getElementById('results').innerHTML = output + '</tbody></table>';
}


function updatePage(output) {
    document.getElementById('output').innerHTML += '<br>' + output;
}


function lastNDays(n) {
    var today = new Date();
    var before = new Date();
    before.setDate(today.getDate() - n);

    var year = before.getFullYear();

    var month = before.getMonth() + 1;
    if (month < 10) {
        month = '0' + month;
    }

    var day = before.getDate();
    if (day < 10) {
        day = '0' + day;
    }

    return [year, month, day].join('-');
}

function createCSV(data){
    var csvContent = "data:text/csv;charset=utf-8,";
    data.forEach(function(infoArray, index){
        dataString = infoArray.join(",");
        csvContent += index < infoArray.length ? dataString+ "\n" : dataString;
    }); 
    var encodedUri = encodeURI(csvContent);
    var link = document.createElement("a");
    link.setAttribute("class", 'btn');
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", "my_data.csv");
    document.getElementById('output').style.display = 'none';
    link.click();
}
