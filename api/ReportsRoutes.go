package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Bind binds the data for the HTTP
func (data *Report) Bind(r *http.Request) error {
	return nil
}

// GetReportReasonsRoute gets the report reasons
func GetReportReasonsRoute(w http.ResponseWriter, r *http.Request) {
	Send(w, http.StatusOK, GetReportReasons())
	return
}

// GetReportStatusesRoute gets the report statuses
func GetReportStatusesRoute(w http.ResponseWriter, r *http.Request) {
	Send(w, http.StatusOK, GetReportStatuses())
	return
}

// ReportRequestRoute allows a request to be reported. Note that the user can be blank; we cannot require users to be logged in
// to report (for example, copyright violations, etc)
func ReportRequestRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, _ := CheckForUser(r)

	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	report := Report{}
	render.Bind(r, &report)

	report.ReporterID = jwtUser.ID
	report.RequestID = requestID
	report.Reason = strings.ToLower(report.Reason)
	report.ReasonText, _ = sanitize(report.ReasonText)
	report.Status = ReportStatusOpen

	validReason := false
	for i := range reportReasons {
		if report.Reason == reportReasons[i] {
			validReason = true
		}
	}

	if !validReason {
		SendError(w, http.StatusBadRequest, "report_reason_invalid", "report reason is invalid", report)
		return
	}

	err := CreateReport(&report)
	if err != nil {
		SendError(w, http.StatusBadRequest, "report_create_failed", "report could not be created", err)
		return
	}
	Send(w, http.StatusCreated, report)
	return
}

// GetReportsOnRequestRoute gets the reports for a single request
func GetReportsOnRequestRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 || jwtUser.PlatformRole != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	reports, err := GetReportsForRequest(requestID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "reports_admin_get_failure", "could not fetch reports for that request", err)
		return
	}

	Send(w, http.StatusOK, reports)
	return
}

// GetReportsOnPlatformRoute gets the reports for the platform by status
func GetReportsOnPlatformRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 || jwtUser.PlatformRole != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = ReportStatusOpen
	}
	reports, _ := GetReportsForPlatform(status)
	Send(w, http.StatusOK, reports)
	return
}

// GetReportRoute gets the report
func GetReportRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 || jwtUser.PlatformRole != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	reportID, reportIDErr := strconv.ParseInt(chi.URLParam(r, "reportID"), 10, 64)
	if reportIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	report, err := GetReport(reportID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	Send(w, http.StatusOK, report)
	return
}

// UpdateReportRoute updates the status of the report
func UpdateReportRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 || jwtUser.PlatformRole != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	reportID, reportIDErr := strconv.ParseInt(chi.URLParam(r, "reportID"), 10, 64)
	if reportIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	report := Report{}
	render.Bind(r, &report)
	if report.Status == "" {
		SendError(w, http.StatusBadRequest, "report_update_status_invalid", "that status is invalid for that report", report)
		return
	}

	found, err := GetReport(reportID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	found.Status = report.Status

	err = UpdateReport(&found)
	if err != nil {
		SendError(w, http.StatusBadRequest, "report_update_error", "could not update that report", report)
		return
	}
	Send(w, http.StatusOK, report)
	return
}
