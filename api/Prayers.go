package api

import (
	"strings"
	"time"
)

const (
	// PrayerRequestStatusPending represents a prayer that is pending a solution
	PrayerRequestStatusPending = "pending"
	// PrayerRequestStatusAnswered represents a prayer that has been answered
	PrayerRequestStatusAnswered = "answered"
	// PrayerRequestStatusNotAnswered represnts a prayer that was not answered
	PrayerRequestStatusNotAnswered = "not_answered"
	// PrayerRequestSatusUnknown represents a prayer that is unknown
	PrayerRequestSatusUnknown = "unknown"

	// PrayerTimeoutInMinutes is the number of minutes between times a user is allowed to submit a prayer made towards a prayer request
	PrayerTimeoutInMinutes = 60 * 6
)

// PrayerRequest represents a single prayer request
type PrayerRequest struct {
	ID          int64    `json:"id" db:"id"`
	Title       string   `json:"title" db:"title"`
	Body        string   `json:"body" db:"body"`
	CreatedBy   int64    `json:"createdBy" db:"createdBy"`
	Privacy     string   `json:"privacy" db:"privacy"`
	DateCreated string   `json:"dateCreated" db:"dateCreated"`
	Status      string   `json:"status" db:"status"`
	Tags        []string `json:"tags" db:"-"`
	PrayerCount int      `json:"prayerCount" db:"prayerCount"`
}

// Prayer represents a prayer made towards a request
type Prayer struct {
	PrayerRequestID int64  `json:"prayerRequestId" db:"prayerRequestId"`
	UserID          int64  `json:"userId" db:"userId"`
	WhenPrayed      string `json:"whenPrayed" db:"whenPrayed"`
}

// PrayerRequestCommunityLink joins a prayer request to a community
type PrayerRequestCommunityLink struct {
	PrayerRequestID int64 `json:"prayerRequestId" db:"prayerRequestId"`
	CommunityID     int64 `json:"communityId" db:"communityId"`
}

// CreatePrayerRequest creates a new prayer request
func CreatePrayerRequest(input *PrayerRequest) error {
	input.processForDB()
	defer input.processForAPI()
	query := `INSERT INTO PrayerRequests (title, body, createdBy, privacy, status, dateCreated) 
		VALUES (:title, :body, :createdBy, :privacy, :status, NOW())`
	res, err := Config.DbConn.NamedExec(query, &input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdatePrayerRequest updates only the privacy or status of the prayer request to avoid things like request editing to make things look awkward
func UpdatePrayerRequest(input *PrayerRequest) error {
	_, err := Config.DbConn.NamedExec("UPDATE PrayerRequests SET privacy = :privacy, status = :status WHERE id = :id LIMIT 1", input)
	return err
}

// DeletePrayerRequest deletes a prayer request, any prayers made for it, and any tag links; a cron will cleanup orphaned tags
func DeletePrayerRequest(id int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM PrayerRequests WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM Prayers WHERE prayerRequestId = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM PrayerRequestCommunityLinks WHERE prayerRequestId = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM PrayerRequestTagLinks WHERE prayerRequestId = ?", id)
	if err != nil {
		return err
	}

	return nil
}

// GetPrayerRequest gets a single prayer request
func GetPrayerRequest(id int64) (*PrayerRequest, error) {
	request := PrayerRequest{}
	err := Config.DbConn.Get(&request, `SELECT pr.*, (SELECT COUNT(*) FROM Prayers p WHERE p.prayerRequestId = pr.id) AS prayerCount 
		FROM PrayerRequests pr WHERE pr.id = ? LIMIT 1`, id)
	request.processForAPI()
	return &request, err
}

// GetUserPrayerRequests gets the requests submitted by a user
func GetUserPrayerRequests(userID int64, status string, startDate, endDate string, count, offset int) ([]PrayerRequest, error) {
	// two queries depending on whether status is passed in
	requests := []PrayerRequest{}
	var err error
	status = strings.ToLower(status)

	// if the dates are blank, assume all time
	if startDate == "" {
		startDate = "2000-01-01 00:00:00"
	}
	if endDate == "" {
		endDate = "3000-01-01 00:00:00"
	}
	if status == "pending" || status == "answered" || status == "not_answered" || status == "unknown" {
		query := `SELECT pr.*, (SELECT COUNT(*) FROM Prayers p WHERE p.prayerRequestId = pr.id) AS prayerCount 
		FROM PrayerRequests pr 
		WHERE createdBy = ? AND status = ? AND dateCreated BETWEEN ? AND ? ORDER BY dateCreated DESC LIMIT ?,?`
		err = Config.DbConn.Select(&requests, query, userID, status, startDate, endDate, offset, count)
	} else {
		query := `SELECT pr.*, (SELECT COUNT(*) FROM Prayers p WHERE p.prayerRequestId = pr.id) AS prayerCount
		FROM PrayerRequests pr 
		WHERE createdBy = ? AND dateCreated BETWEEN ? AND ? ORDER BY dateCreated DESC LIMIT ?,?`
		err = Config.DbConn.Select(&requests, query, userID, startDate, endDate, offset, count)
	}
	for i := range requests {
		requests[i].processForAPI()
	}
	return requests, err
}

// AddPrayerMade adds a prayer made towards a request
func AddPrayerMade(userID, prayerRequestID int64) error {
	when := time.Now().Format("2006-01-02 15:04:05")
	_, err := Config.DbConn.Exec("INSERT INTO Prayers (userId, prayerRequestId, whenPrayed) VALUES (?,?,?)",
		userID, prayerRequestID, when)
	return err
}

// RemovePrayerMade removes a prayer made towards a request
func RemovePrayerMade(userID, prayerRequestID int64, whenPrayed string) error {
	_, err := Config.DbConn.Exec("DELETE FROM Prayers WHERE userId = ? AND prayerRequestId = ? AND whenPrayed = ?",
		userID, prayerRequestID, whenPrayed)
	return err
}

// GetPrayersMadeByUserForRequest gets the prayers made for a specific user and request, ordered by whenPrayed DESC
func GetPrayersMadeByUserForRequest(userID, prayerRequestID int64, count, offset int) ([]Prayer, error) {
	prayers := []Prayer{}
	err := Config.DbConn.Select(&prayers, "SELECT * FROM Prayers WHERE userId = ? AND prayerRequestId = ? ORDER BY whenPrayed DESC LIMIT ?,?", userID, prayerRequestID, offset, count)
	return prayers, err
}

// CanUserMakeNewPrayer checks to see if the user is allowed to submit another prayer made for a request
func CanUserMakeNewPrayer(userID, prayerRequestID int64) (canSubmitPrayer bool, minutesUntilNextAllowed float64) {
	// a user can make one prayer per request every calendar day
	// first, has the user made a request before?
	prayer := Prayer{}
	canSubmitPrayer = false
	prayers, err := GetPrayersMadeByUserForRequest(userID, prayerRequestID, 1, 0)
	if err != nil || len(prayers) == 0 {
		// no prayers
		return true, 0
	}
	prayer = prayers[0]
	// they have submitted one in the past, so find out if enough time has passed
	whenPrayed, _ := time.Parse("2006-01-02 15:04:05", prayer.WhenPrayed)
	nextAllowed := whenPrayed.Add(time.Minute * PrayerTimeoutInMinutes)

	if nextAllowed.Before(time.Now()) {
		// allowed
		return true, 0
	}

	// find out the difference
	allowedIn := nextAllowed.Sub(time.Now()).Minutes()
	return false, allowedIn
}

// AddPrayerRequestToCommunity adds a request to a community; 0 means it is on the public feed
func AddPrayerRequestToCommunity(requestID, communityID int64) error {
	_, err := Config.DbConn.Exec(`INSERT INTO PrayerRequestCommunityLinks (prayerRequestId, communityId) VALUES (?,?) ON DUPLICATE KEY UPDATE communityId = ?`, requestID, communityID, communityID)
	return err
}

// RemovePrayerRequestFromCommunity removes a request from a community
func RemovePrayerRequestFromCommunity(requestID, communityID int64) error {
	_, err := Config.DbConn.Exec(`DELETE FROM PrayerRequestCommunityLinks WHERE prayerRequestId = ? AND communityId = ?`, requestID, communityID)
	return err
}

// GetCountOfRequestsInCommunity gets the number of requests in a community within a time frame
func GetCountOfRequestsInCommunity(communityID int64, start, end string) (int64, error) {
	count := struct {
		Count int64 `db:"count"`
	}{}
	if start == "" {
		start = "1970-01-01 00:00:00"
	}
	if end == "" {
		end = "2100-01-01 00:00:00"
	}
	err := Config.DbConn.Get(&count, `SELECT COUNT(*) AS count FROM PrayerRequestCommunityLinks prcl, PrayerRequests pr WHERE prcl.communityId = ? AND prcl.prayerRequestId = pr.id AND pr.dateCreated BETWEEN ? AND ?`,
		communityID, start, end)
	return count.Count, err
}

// GetCountOfPrayersMadeForRequest gets the number of unique prayers for a specific request
func GetCountOfPrayersMadeForRequest(requestID int64, start, end string) (count int64) {
	countStruct := struct {
		Count int64 `db:"count"`
	}{}
	if start != "" && end != "" {
		Config.DbConn.Get(&countStruct, "SELECT COUNT(*) AS count FROM Prayers WHERE prayerRequestId = ? AND whenPrayed BETWEEN ? AND ?", requestID, start, end)
	} else {
		Config.DbConn.Get(&countStruct, "SELECT COUNT(*) AS count FROM Prayers WHERE prayerRequestId = ?", requestID)
	}
	return countStruct.Count
}

// processForDB ensures data consistency
func (u *PrayerRequest) processForDB() {

	if u.DateCreated == "" {
		u.DateCreated = "1970-01-01"
	} else {
		parsed, err := time.Parse("2006-01-02T15:04:05Z", u.DateCreated)
		if err != nil {
			// perhaps it was already db time
			parsed, err = time.Parse("2006-01-02 15:04:05", u.DateCreated)
			if err != nil {
				// last try; no time?
				parsed, err = time.Parse("2006-01-02", u.DateCreated)
				if err != nil {
					// screw it
					parsed, _ = time.Parse("2006-01-02", "1970-01-01")
				}
			}
		}
		u.DateCreated = parsed.Format("2006-01-02")
	}

	if u.Status == "" {
		u.Status = "pending"
	}
}

// processForAPI ensures data consistency and creates the JWT
func (u *PrayerRequest) processForAPI() {
	if u == nil {
		return
	}

	if u.DateCreated == "1970-01-01 00:00:00" {
		u.DateCreated = ""
	} else {
		u.DateCreated, _ = ParseTimeToDate(u.DateCreated)
	}

	if u.Status == "" {
		u.Status = "pending"
	}

}
