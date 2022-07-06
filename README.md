# CDC-Covid

Quick program to grab ONLY the JSON file from the CDC Covid-19 site and parse it for the latest reported covid levels for a given FIPS Code (County Id)

***NOTE***: This WILL break if they change the JSON format or change the query arguments for the JSON file as it's not a public API end-point!

## Installation

```bash
git clone https://github.com/zariok/cdc-covid
cd cdc-covid
go build
```

If you want to be notified via text message, create a ```twilio.env``` file as follows with your appropriate information:

```bash
export TWILIO_ACCOUNT_SID='abcdef'
export TWILIO_AUTH_TOKEN='abcdef'
export TWILIO_PHONE_FROM='+19994443333'
export TWILIO_PHONE_TO='+19995551000,+19995552000'
``` 

## How to find FIPS Code

Navigate to [Covid-19 Integrated County View](https://covid.cdc.gov/covid-data-tracker/#county-view) and select your state and county.

Look at the URL, or select it if it's "hidden" and look for:

***list_select_county=12095*** where 12095 is the FIPS Code for Orange County, Florida


## Running

```bash
./cdc-covid -id 12095
```

A simple bash script, which could be cron'd:
```bash
#!/bin/bash
cd /home/HOMEDIR/cdc-covid
source ./twilio.env
./cdc-covid -id 12095
```



### License

[MIT](https://choosealicense.com/licenses/mit/) 
