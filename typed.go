package main

import (
	"encoding/json"
	"strings"
)

type Manifestv2022_10 struct {
	Shared   []string `json:"shared"`
	Ondemand struct {
		Cv      []string `json:"cv"`
		Cv2204  []string `json:"cv2204"`
		Dulv2   []string `json:"dulv2"`
		Dulv3   []string `json:"dulv3"`
		Dulv4   []string `json:"dulv4"`
		Duhwocr []string `json:"duhwocr"`
	} `json:"ondemand"`
	Actioncenter                  []string `json:"actioncenter"`
	Aicenter                      []string `json:"aicenter"`
	AuditService                  []string `json:"audit-service"`
	Automationhub                 []string `json:"automationhub"`
	BusinessApps                  []string `json:"business-apps"`
	DocumentUnderstanding         []string `json:"document-understanding"`
	Identity                      []string `json:"identity"`
	Insights                      []string `json:"insights"`
	LicenseAccountant             []string `json:"license-accountant"`
	LicenseResourceManager        []string `json:"license-resource-manager"`
	LocationService               []string `json:"location-service"`
	Orchestrator                  []string `json:"orchestrator"`
	OrganizationManagementService []string `json:"organization-management-service"`
	Portal                        []string `json:"portal"`
	StudioGovernance              []string `json:"studio-governance"`
	Taskmining                    []string `json:"taskmining"`
	Testmanager                   []string `json:"testmanager"`
	Webhook                       []string `json:"webhook"`
	Dataservice                   []string `json:"dataservice"`
	ResourceCatalogService        []string `json:"resource-catalog-service"`
	ProcessMining                 []string `json:"process-mining"`
}

func getTagListFrom2022_10(body []byte) ([]string, error) {
	var v202210 Manifestv2022_10
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&v202210)
	if err != nil {
		return nil, err
	}

	var result []string
	result = append(result, v202210.Shared...)
	result = append(result, v202210.Actioncenter...)
	result = append(result, v202210.Aicenter...)
	result = append(result, v202210.AuditService...)
	result = append(result, v202210.Automationhub...)
	result = append(result, v202210.BusinessApps...)
	result = append(result, v202210.DocumentUnderstanding...)
	result = append(result, v202210.Identity...)
	result = append(result, v202210.Insights...)
	result = append(result, v202210.LicenseAccountant...)
	result = append(result, v202210.LicenseResourceManager...)
	result = append(result, v202210.LocationService...)
	result = append(result, v202210.Orchestrator...)
	result = append(result, v202210.OrganizationManagementService...)
	result = append(result, v202210.Taskmining...)
	result = append(result, v202210.Webhook...)
	result = append(result, v202210.StudioGovernance...)
	result = append(result, v202210.Testmanager...)
	result = append(result, v202210.Portal...)
	return result, nil
}

func getTypedTagList(body []byte) ([]string, error) {
	result, err := getTagListFrom2022_10(body)
	if err == nil {
		return result, nil
	}

	return []string{}, nil
}
