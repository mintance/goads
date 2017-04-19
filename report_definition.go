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
	"time"
)

type ReportsRow struct {
	Clicks string `json:"-clicks"`
	CampaignID string `json:"-campaignID"`
	AdGroupID string `json:"-adGroupID"`
	AdID string `json:"-adID"`
	Impressions string `json:"-impressions"`
	Cost string `json:"-cost"`
	Currency string `json:"-currency"`
	Headline string `json:"-headline1"`
	Date string `json:"-day"`
	KeywordId string `json:"-keywordID"`
	KeywordName string
}

type AdReports struct {
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

type KeywordReportsRow struct {
	KeywordName string `json:"-keywordPlacement"`
	KeywordId string `json:"-keywordID"`
	Date string `json:"-day"`
}

type KeywordReports struct {
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
			Rows []KeywordReportsRow `json:"row"`
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

	today := time.Now().Local().Format("20060102")

	yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")

	adReports := []byte(`__rdxml=<?xml version="1.0" encoding="UTF-8"?>
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
					<fields>CriterionId</fields>
					<fields>HeadlinePart1</fields>
					<dateRange>
					  <min>`+yesterday+`</min>
					  <max>`+today+`</max>
					</dateRange>
				</selector>
				<reportName>Mintance</reportName>
				<reportType>AD_PERFORMANCE_REPORT</reportType>
				<dateRangeType>CUSTOM_DATE</dateRangeType>
				<downloadFormat>XML</downloadFormat>
				</reportDefinition>`)

	adKeywords := []byte(`__rdxml=<?xml version="1.0" encoding="UTF-8"?>
				<reportDefinition>
				<selector>
					<fields>Id</fields>
					<fields>Criteria</fields>
					<fields>Date</fields>
					<dateRange>
					  <min>`+yesterday+`</min>
					  <max>`+today+`</max>
					</dateRange>
				</selector>
				<reportName>Mintance Keyword</reportName>
				<reportType>CRITERIA_PERFORMANCE_REPORT</reportType>
				<dateRangeType>CUSTOM_DATE</dateRangeType>
				<downloadFormat>XML</downloadFormat>
				</reportDefinition>`)

	keywords := s.requestAdKeywords(adKeywords)

	adverts := s.requestAdReport(adReports)

	for _, keyword := range keywords{

		for key := range adverts {
			if keyword.KeywordId == adverts[key].KeywordId || keyword.Date == adverts[key].Date {
				adverts[key].KeywordName = keyword.KeywordName
			}
		}
	}
	return adverts
}

func (s ReportDefinitionService) requestAdReport(reqBody []byte) []ReportsRow{

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
	reports := AdReports{}

	json.Unmarshal([]byte(reportsJson.String()), &reports)

	return  reports.Report.Table.Rows
}

func (s ReportDefinitionService) requestAdKeywords(reqBody []byte) []KeywordReportsRow{

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
	reports := KeywordReports{}

	json.Unmarshal([]byte(reportsJson.String()), &reports)

	return reports.Report.Table.Rows

}
