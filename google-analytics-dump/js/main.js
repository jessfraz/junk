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
            'metrics': 'ga:pageviews, ga:visits',
            'dimensions': 'ga:date, ga:pageTitle',
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
