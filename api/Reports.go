package api

import "time"

// Report is a report on a request
type Report struct {
	ID           int64  `json:"id" db:"id"`
	RequestID    int64  `json:"requestId" db:"requestId"`
	ReporterID   int64  `json:"reporterId" db:"reporterId"`
	Reason       string `json:"reason" db:"reason"`
	ReasonText   string `json:"reasonText" db:"reasonText"`
	ReportedOn   string `json:"reportedOn" db:"reportedOn"`
	LastUpdated  string `json:"lastUpdated" db:"lastUpdated"`
	Status       string `json:"status" db:"status"`
	RequestTitle string `json:"requestTitle" db:"requestTitle"`
}

const (
	// ReportStatusOpen is a report that is open
	ReportStatusOpen = "open"
	// ReportStatusClosedNoAction is a report that is closed without any additional action
	ReportStatusClosedNoAction = "closed_no_action"
	// ReportStatusClosedDeleted is a report that is closed and the request was removed
	ReportStatusClosedDeleted = "closed_deleted"
	// ReportStatusFollowUp is a report that needs a follow up or research
	ReportStatusFollowUp = "follow_up"

	// ReportReasonThreat is a report that the request is threatening
	ReportReasonThreat = "threat"
	// ReportReasonOffensive is a report that the request is offensive
	ReportReasonOffensive = "offensive"
	// ReportReasonCopyright is a report that the request is a copyright violation
	ReportReasonCopyright = "copyright"
	// ReportReasonOther is a report that the request is unknown or other
	ReportReasonOther = "other"
)

var reportStatuses = []string{ReportStatusOpen, ReportStatusClosedNoAction, ReportStatusClosedDeleted, ReportStatusFollowUp}
var reportReasons = []string{ReportReasonThreat, ReportReasonOffensive, ReportReasonCopyright, ReportReasonOther}

// GetReportStatuses gets the report statuses
func GetReportStatuses() []string {
	return reportStatuses
}

// GetReportReasons gets the report reasons
func GetReportReasons() []string {
	return reportReasons
}

// CreateReport adds a report for a request
func CreateReport(input *Report) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := Config.DbConn.NamedExec(`INSERT INTO Reports (requestId, reporterId, reason, reasonText, reportedOn, lastUpdated, status) 
		VALUES (:requestId, :reporterId, :reason, :reasonText, NOW(), NOW(), :status)`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateReport updates a report's status
func UpdateReport(input *Report) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := Config.DbConn.NamedExec(`UPDATE Reports SET status = :status, lastUpdated = NOW() WHERE id = :id`, input)
	return err
}

// DeleteReportForTest deletes a report for a test; in real use, the status should be updated
func DeleteReportForTest(reportID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM Reports WHERE id = ?", reportID)
	return err
}

// GetReport gets a single report
func GetReport(reportID int64) (Report, error) {
	report := Report{}
	err := Config.DbConn.Get(&report, "SELECT r.*, pr.title AS requestTitle FROM Reports r, PrayerRequests pr WHERE r.id = ? AND r.requestId = pr.id", reportID)
	report.processForAPI()
	return report, err
}

// GetReportsForRequest gets all reports for a request
func GetReportsForRequest(requestID int64) ([]Report, error) {
	reports := []Report{}
	err := Config.DbConn.Select(&reports, "SELECT r.*, pr.title AS requestTitle FROM Reports r, PrayerRequests pr WHERE r.requestId = ? AND r.requestId = pr.id", requestID)
	for i := range reports {
		reports[i].processForAPI()
	}
	return reports, err
}

// GetReportsForPlatform gets all of the reports for the platform
func GetReportsForPlatform(status string) ([]Report, error) {
	reports := []Report{}
	err := Config.DbConn.Select(&reports, "SELECT r.*, pr.title AS requestTitle FROM Reports r, PrayerRequests pr WHERE r.status = ? AND r.requestId = pr.id", status)
	for i := range reports {
		reports[i].processForAPI()
	}
	return reports, err
}

func (u *Report) processForDB() {

	if u.ReportedOn == "" {
		u.ReportedOn = time.Now().Format("2006-01-02 15:04:05")
	}

	if u.Status == "" {
		u.Status = ReportStatusOpen
	}
}

func (u *Report) processForAPI() {

	if u.ReportedOn == "" {
		u.ReportedOn = time.Now().Format("2006-01-02T15:04:05Z")
	} else {
		parsed, _ := time.Parse("2006-01-02 15:04:05", u.ReportedOn)
		u.ReportedOn = parsed.Format("2006-01-02T15:04:05Z")
	}

	if u.Status == "" {
		u.Status = ReportStatusOpen
	}
}
