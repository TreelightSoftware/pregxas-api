package api

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportVariables(t *testing.T) {
	statuses := GetReportStatuses()
	assert.NotZero(t, len(statuses))
	assert.Equal(t, len(reportStatuses), len(statuses))

	reasons := GetReportReasons()
	assert.NotZero(t, len(reasons))
	assert.Equal(t, len(reportReasons), len(reasons))
}

func TestReportsCrud(t *testing.T) {
	ConfigSetup()
	r := rand.Int63n(99999999)

	request := PrayerRequest{
		Title:     "Bad Request",
		Body:      "Bad",
		CreatedBy: r,
	}
	err := CreatePrayerRequest(&request)
	require.Nil(t, err)
	require.NotZero(t, request.ID)
	defer DeletePrayerRequest(request.ID)

	report := Report{
		RequestID:  request.ID,
		ReporterID: r,
		Reason:     ReportReasonThreat,
		ReasonText: "This is offensive",
	}
	err = CreateReport(&report)
	assert.Nil(t, err)
	require.NotZero(t, report.ID)
	defer DeleteReportForTest(report.ID)
	assert.Equal(t, ReportStatusOpen, report.Status)

	// get it in a variety of ways
	found, err := GetReport(report.ID)
	assert.Nil(t, err)
	assert.Equal(t, report.ID, found.ID)
	assert.Equal(t, report.RequestID, found.RequestID)
	assert.Equal(t, report.ReporterID, found.ReporterID)
	assert.Equal(t, ReportStatusOpen, found.Status)
	assert.Equal(t, report.Reason, found.Reason)
	assert.Equal(t, report.ReasonText, found.ReasonText)
	assert.Equal(t, request.Title, found.RequestTitle)

	foundForRequest, err := GetReportsForRequest(request.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundForRequest))
	found = foundForRequest[0]
	assert.Equal(t, report.ID, found.ID)
	assert.Equal(t, report.RequestID, found.RequestID)
	assert.Equal(t, report.ReporterID, found.ReporterID)
	assert.Equal(t, ReportStatusOpen, found.Status)
	assert.Equal(t, report.Reason, found.Reason)
	assert.Equal(t, report.ReasonText, found.ReasonText)
	assert.Equal(t, request.Title, found.RequestTitle)

	foundOnPlatform, err := GetReportsForPlatform(ReportStatusClosedNoAction)
	assert.Nil(t, err)
	foundInLoop := false
	for i := range foundOnPlatform {
		if foundOnPlatform[i].ID == report.ID {
			foundInLoop = true
		}
	}
	assert.False(t, foundInLoop)

	foundOnPlatform, err = GetReportsForPlatform(ReportStatusOpen)
	assert.Nil(t, err)
	foundInLoop = false
	for i := range foundOnPlatform {
		if foundOnPlatform[i].ID == report.ID {
			foundInLoop = true
		}
	}
	assert.True(t, foundInLoop)

	// update it and get it again
	report.Status = ReportStatusClosedNoAction
	err = UpdateReport(&report)
	assert.Nil(t, err)

	found, err = GetReport(report.ID)
	assert.Nil(t, err)
	assert.Equal(t, report.ID, found.ID)
	assert.Equal(t, report.RequestID, found.RequestID)
	assert.Equal(t, report.ReporterID, found.ReporterID)
	assert.Equal(t, ReportStatusClosedNoAction, found.Status)
	assert.Equal(t, report.Reason, found.Reason)
	assert.Equal(t, report.ReasonText, found.ReasonText)
	assert.Equal(t, request.Title, found.RequestTitle)

	foundForRequest, err = GetReportsForRequest(request.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundForRequest))
	found = foundForRequest[0]
	assert.Equal(t, report.ID, found.ID)
	assert.Equal(t, report.RequestID, found.RequestID)
	assert.Equal(t, report.ReporterID, found.ReporterID)
	assert.Equal(t, ReportStatusClosedNoAction, found.Status)
	assert.Equal(t, report.Reason, found.Reason)
	assert.Equal(t, report.ReasonText, found.ReasonText)
	assert.Equal(t, request.Title, found.RequestTitle)

	foundOnPlatform, err = GetReportsForPlatform(ReportStatusClosedNoAction)
	assert.Nil(t, err)
	foundInLoop = false
	for i := range foundOnPlatform {
		if foundOnPlatform[i].ID == report.ID {
			foundInLoop = true
		}
	}
	assert.True(t, foundInLoop)

	foundOnPlatform, err = GetReportsForPlatform(ReportStatusOpen)
	assert.Nil(t, err)
	foundInLoop = false
	for i := range foundOnPlatform {
		if foundOnPlatform[i].ID == report.ID {
			foundInLoop = true
		}
	}
	assert.False(t, foundInLoop)
}
