package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsCRUDRoutes(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999999)
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	admin := User{
		PlatformRole: "admin",
	}
	err := CreateTestUser(&admin)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&admin)

	user := User{}
	err = CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user)

	request := PrayerRequest{
		Title:     fmt.Sprintf("Test Prayer %d", randID),
		Body:      "Test Prayer Request Body",
		CreatedBy: user.ID,
		Privacy:   "public",
	}
	err = CreatePrayerRequest(&request)
	require.Nil(t, err)
	defer DeletePrayerRequest(request.ID)

	// first, get the statuses and the reasons

	code, res, _ := TestAPICall(http.MethodGet, "/admin/reports/reasons", b, GetReportReasonsRoute, "", "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ := UnmarshalTestArray(res)
	assert.Equal(t, len(GetReportReasons()), len(bodyA))

	code, res, _ = TestAPICall(http.MethodGet, "/admin/reports/statuses", b, GetReportStatusesRoute, "", "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Equal(t, len(GetReportStatuses()), len(bodyA))

	// now create two reports (one anonymous and one with a user) and get them with a variety of inputs
	input1 := Report{
		RequestID:  request.ID,
		ReasonText: "This report is offensive to me",
	}

	input2 := Report{
		RequestID:  request.ID,
		Reason:     ReportReasonThreat,
		ReasonText: "This report is threatening to me",
	}

	// make a bad call
	code, res, _ = TestAPICall(http.MethodPost, "/requests/a/reports", b, ReportRequestRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)

	// make a bad reason
	b.Reset()
	enc.Encode(&input1)
	code, res, _ = TestAPICall(http.MethodPost, fmt.Sprintf("/requests/%d/reports", request.ID), b, ReportRequestRoute, "", "")
	require.Equal(t, http.StatusBadRequest, code)

	input1.Reason = ReportReasonOffensive
	b.Reset()
	enc.Encode(&input1)
	code, res, _ = TestAPICall(http.MethodPost, fmt.Sprintf("/requests/%d/reports", request.ID), b, ReportRequestRoute, "", "")
	require.Equal(t, http.StatusCreated, code)

	_, body, _ := UnmarshalTestMap(res)
	report1 := Report{}
	mapstructure.Decode(body, &report1)
	assert.NotZero(t, report1.ID)
	assert.Equal(t, input1.Reason, report1.Reason)
	assert.Equal(t, input1.ReasonText, report1.ReasonText)
	defer DeleteReportForTest(report1.ID)

	b.Reset()
	enc.Encode(&input2)
	code, res, _ = TestAPICall(http.MethodPost, fmt.Sprintf("/requests/%d/reports", request.ID), b, ReportRequestRoute, user.JWT, "")
	require.Equal(t, http.StatusCreated, code)
	_, body, _ = UnmarshalTestMap(res)
	report2 := Report{}
	mapstructure.Decode(body, &report2)
	assert.NotZero(t, report2.ID)
	assert.Equal(t, input2.Reason, report2.Reason)
	assert.Equal(t, input2.ReasonText, report2.ReasonText)
	defer DeleteReportForTest(report2.ID)

	// get them

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report1.ID), b, GetReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	found1 := Report{}
	mapstructure.Decode(body, &found1)
	assert.NotZero(t, found1.ID)
	assert.Equal(t, input1.Reason, found1.Reason)
	assert.Equal(t, input1.ReasonText, found1.ReasonText)
	assert.Equal(t, ReportStatusOpen, found1.Status)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report2.ID), b, GetReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	found2 := Report{}
	mapstructure.Decode(body, &found2)
	assert.NotZero(t, found2.ID)
	assert.Equal(t, input2.Reason, found2.Reason)
	assert.Equal(t, input2.ReasonText, found2.ReasonText)
	assert.Equal(t, ReportStatusOpen, found2.Status)

	code, _, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report1.ID), b, GetReportRoute, user.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	code, _, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report2.ID), b, GetReportRoute, user.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	// update them

	// first, poorly

	code, res, _ = TestAPICall(http.MethodPatch, "/admin/reports/a", b, UpdateReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	badUpdate := Report{
		Status: "notreal",
	}
	b.Reset()
	enc.Encode(&badUpdate)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/admin/reports/%d", report1.ID), b, UpdateReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusBadRequest, code)

	update := Report{
		Status: ReportStatusClosedNoAction,
	}
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, "/admin/reports/0", b, UpdateReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/admin/reports/%d", report1.ID), b, UpdateReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/admin/reports/%d", report1.ID), b, UpdateReportRoute, user.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	update2 := Report{
		Status: ReportStatusClosedDeleted,
	}
	b.Reset()
	enc.Encode(&update2)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/admin/reports/%d", report2.ID), b, UpdateReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/admin/reports/%d", report2.ID), b, UpdateReportRoute, user.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	// get them again
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report1.ID), b, GetReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	found1 = Report{}
	mapstructure.Decode(body, &found1)
	assert.NotZero(t, found1.ID)
	assert.Equal(t, input1.Reason, found1.Reason)
	assert.Equal(t, input1.ReasonText, found1.ReasonText)
	assert.Equal(t, ReportStatusClosedNoAction, found1.Status)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/admin/reports/%d", report2.ID), b, GetReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	found2 = Report{}
	mapstructure.Decode(body, &found2)
	assert.NotZero(t, found2.ID)
	assert.Equal(t, input2.Reason, found2.Reason)
	assert.Equal(t, input2.ReasonText, found2.ReasonText)
	assert.Equal(t, ReportStatusClosedDeleted, found2.Status)

	// get the reports on a request and the platform

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/requests/%d/reports", request.ID), b, GetReportRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	foundInList1 := false
	foundInList2 := false
	for i := range bodyA {
		fR := Report{}
		mapstructure.Decode(bodyA[i], &fR)
		if fR.ID == report1.ID {
			foundInList1 = true
		}
		if fR.ID == report2.ID {
			foundInList2 = true
		}
	}
	assert.True(t, foundInList1)
	assert.True(t, foundInList2)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("admin/reports?status=%s", ReportStatusClosedDeleted), b, GetReportsOnPlatformRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	foundInList1 = false
	foundInList2 = false
	for i := range bodyA {
		fR := Report{}
		mapstructure.Decode(bodyA[i], &fR)
		if fR.ID == report1.ID {
			foundInList1 = true
		}
		if fR.ID == report2.ID {
			foundInList2 = true
		}
	}
	assert.False(t, foundInList1)
	assert.True(t, foundInList2)
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("admin/reports?status=%s", ReportStatusClosedNoAction), b, GetReportsOnPlatformRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	foundInList1 = false
	foundInList2 = false
	for i := range bodyA {
		fR := Report{}
		mapstructure.Decode(bodyA[i], &fR)
		if fR.ID == report1.ID {
			foundInList1 = true
		}
		if fR.ID == report2.ID {
			foundInList2 = true
		}
	}
	assert.True(t, foundInList1)
	assert.False(t, foundInList2)
}
