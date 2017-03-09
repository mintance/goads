package gads

import (
	"fmt"
	"bytes"
	"io/ioutil"
	"strings"
	"encoding/json"
	"net/http"
	xj "github.com/basgys/goxml2json"
	"bitbucket.org/kargell_marketing_backend/mintance.core"
	"github.com/Sirupsen/logrus"
)

type ReportsRow struct {
	Clicks string `json:"-clicks"`
	CampaignID string `json:"-campaignID"`
	AdGroupID string `json:"-adGroupID"`
	AdID string `json:"-adID"`
	Impressions string `json:"-impressions"`
	Cost string `json:"-cost"`
	Currency string `json:"-currency"`
	Date string `json:"-day"`
}

type Reports struct {
	Report struct {
		ReportName struct {
			Name string `json:"-name"`
		} `json:"report-name"`
		DateRange struct {
			Date string `json:"-date"`
		} `json:"date-range"`
		Table struct {
			Columns struct {
				Column []struct {
					Name string `json:"-name"`
					Display string `json:"-display"`
				} `json:"column"`
			} `json:"columns"`
			Rows []ReportsRow `json:"row"`
		} `json:"table"`
	} `json:"report"`
}

type ReportDefinitionService struct {
	AuthConfig
}

func NewReportDefinitionService(auth *AuthConfig) *ReportDefinitionService {
	return &ReportDefinitionService{AuthConfig: *auth}
}


func (s ReportDefinitionService) GetReport() []ReportsRow{

	reqBody := []byte(`__rdxml=<?xml version="1.0" encoding="UTF-8"?>
				<reportDefinition>
				<selector>
					<fields>CampaignId</fields>
					<fields>AdGroupId</fields>
					<fields>Id</fields>
					<fields>Impressions</fields>
					<fields>Clicks</fields>
					<fields>Cost</fields>
					<fields>Date</fields>
					<fields>AccountCurrencyCode</fields>
				</selector>
				<reportName>Mintance</reportName>
				<reportType>AD_PERFORMANCE_REPORT</reportType>
				<dateRangeType>TODAY</dateRangeType>
				<downloadFormat>XML</downloadFormat>
				</reportDefinition>`)

	return s.request(reqBody)
}

func (s ReportDefinitionService) request(reqBody []byte) []ReportsRow{

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://adwords.google.com/api/adwords/reportdownload/v201609", bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	contentLength := fmt.Sprintf("%d", len(reqBody))
	req.Header.Add("Content-length", contentLength)
	req.Header.Add("developerToken", s.Auth.DeveloperToken)
	req.Header.Add("clientCustomerId", s.Auth.CustomerId)
	req.Header.Add("Authorization", "Bearer "+s.GetAccessToken())
	req.Header.Add("Content-Type", "application/xml; charset=utf-8")

	if err != nil {
		mintance.Log(logrus.Fields{
			"CustomerId": s.Auth.CustomerId,
		}, "error","[Reports]: New request error", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		mintance.Log(logrus.Fields{
			"CustomerId": s.Auth.CustomerId,
		}, "error","[Reports]: Do request error", err.Error())	}

	respBody, err := ioutil.ReadAll(resp.Body)

	reportsJson, err := xj.Convert(strings.NewReader(string(respBody)))

	if err != nil {
		mintance.Log(logrus.Fields{
			"CustomerId": s.Auth.CustomerId,
		}, "error","[Reports]: Convert xml to json error", err.Error())
	}
	reports := Reports{}

	json.Unmarshal([]byte(reportsJson.String()), &reports)

	return  reports.Report.Table.Rows
}
