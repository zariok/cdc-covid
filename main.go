package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type LastData struct {
	RunId int64 `json:"runid"`
	Data
}

type Data struct {
	Date                       string  `json:"date"`
	AvgNewCases                string  `json:"new_cases_7_day_rolling_average"`
	AvgNewDeaths               string  `json:"new_deaths_7_day_rolling_average"`
	PercentPositive            float64 `json:"percent_positive_7_day,omitempty"`
	ReportDateStart            string  `json:"report_date_window_start,omitempty"`
	ReportDateEnd              string  `json:"report_date_window_end,omitempty"`
	CommunityTransmissionLevel string  `json:"community_transmission_level,omitempty"`
	CountyName                 string  `json:"county"`
	State                      string  `json:"state"`
	FipsCode                   int64   `json:"fips_code"`
}

type TimeSeriesData struct {
	RunId int64  `json:"runid"`
	Data  []Data `json:"integrated_county_timeseries_external_data"`
}

func main() {
	var flgCountyId = flag.Int64("id", 0, "fips_code county id")
	flag.Parse()

	if *flgCountyId <= 0 {
		fmt.Printf("-id [countyId] is required.\n")
		os.Exit(0)
	}

	countyFile := fmt.Sprintf("lastdata.%d.json", *flgCountyId)

	lastData := LastData{}
	if _, err := os.Stat(countyFile); err == nil {
		jsonFile, err := os.Open(countyFile)
		if err != nil {
			fmt.Printf("unable to open file: %v\n", err)
			os.Exit(1)
		}
		jsonData, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()
		json.Unmarshal(jsonData, &lastData)
	}

	timeSeries := make(map[string]Data)

	url := fmt.Sprintf("https://covid.cdc.gov/covid-data-tracker/COVIDData/getAjaxData?id=integrated_county_timeseries_fips_%d_external", *flgCountyId)
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("no response from %v\n", url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("unable to ready the body response")
	} else {
		var d TimeSeriesData
		err := json.Unmarshal(body, &d)
		if err != nil {
			fmt.Printf("unable to decode json: %v\n", err)
		} else {
			if lastData.RunId == d.RunId {
				fmt.Printf("no new data\n")
				os.Exit(0)
			}
			lastData.RunId = d.RunId
			for _, item := range d.Data {
				timeSeries[item.Date] = item
			}
		}
	}

	/*
		for t, data := range timeSeries {
			log.Printf("%v%% on %v\n", data.PercentPositive, t.Format("2006-02-03"))
		}
	*/
	cntr := 0
	lookTime := time.Now()
	for {
		if cntr > len(timeSeries) {
			fmt.Printf("unable to find data...\n")
			return
		}
		if data, ok := timeSeries[lookTime.Format("2006-01-02")]; ok {
			fmt.Printf("%v %v%% - %v\n", data.Date, data.PercentPositive, data.CommunityTransmissionLevel)
			if data.PercentPositive > 0.0 {
				lastData.Data = data
				break
			}
		} else {
			fmt.Printf("%v NO DATA\n", lookTime.Format("2006-01-02"))
		}
		lookTime = lookTime.AddDate(0, 0, -1)
		cntr++
	}

	// write new file
	file, _ := json.MarshalIndent(lastData, "", " ")
	err = ioutil.WriteFile(countyFile, file, 0644)
	if err != nil {
		fmt.Printf("unable to write the lastdata file: %v\n", err)
	}

	// send text message via twilio (using API)
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if len(accountSid) == 0 || len(authToken) == 0 {
		fmt.Printf("no twilio credentials; TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN\n")
		os.Exit(0)
	}
	phoneFrom := os.Getenv("TWILIO_PHONE_FROM")
	phoneTo := os.Getenv("TWILIO_PHONE_TO")
	if len(phoneFrom) == 0 || len(phoneTo) == 0 {
		fmt.Printf("phone numbers missing; TWILIO_PHONE_FROM, TWILIO_PHONE_TO\n")
		os.Exit(0)
	}

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetFrom(phoneFrom)

	phoneNums := strings.Split(phoneTo, ",")
	for idx := range phoneNums {
		phoneNum := phoneNums[idx]
		params.SetTo(phoneNum)
		params.SetBody(fmt.Sprintf("%v, %v\n%v - %v%% - %v\n", lastData.CountyName, lastData.State, lastData.Date, lastData.PercentPositive, lastData.CommunityTransmissionLevel))

		twilioResp, err := twilioClient.Api.CreateMessage(params)
		if err != nil {
			fmt.Printf("error sending to %s -- %v\n", phoneNum, err.Error())
			err = nil
		} else {
			fmt.Printf("successful msg to phone: %s sid: %s\n", phoneNum, *twilioResp.Sid)
		}
	}
}
