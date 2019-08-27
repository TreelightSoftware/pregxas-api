package api

import (
	"strings"
	"time"
)

// PrayerList is a list used to track prayers
type PrayerList struct {
	ID              int64           `json:"id" db:"id"`
	UserID          int64           `json:"userId" db:"userId"`
	Title           string          `json:"title" db:"title"`
	UpdateFrequency string          `json:"updateFrequency" db:"updateFrequency"`
	Created         string          `json:"created" db:"created"`
	PrayerRequests  []PrayerRequest `json:"prayerRequests"`
}

const (
	// PrayerListUpdateFrequencyDaily instructs the platform to send daily updates about the list
	PrayerListUpdateFrequencyDaily = "daily"
	// PrayerListUpdateFrequencyWeekly instructs the platform to send weekly updates about the list
	PrayerListUpdateFrequencyWeekly = "weekly"
	// PrayerListUpdateFrequencyNever instructs the platform to never send updates about the list
	PrayerListUpdateFrequencyNever = "never"
)

// CreatePrayerList creates a new list for the user
func CreatePrayerList(input *PrayerList) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := Config.DbConn.NamedExec(`INSERT INTO PrayerLists (userId, title, updateFrequency, created) VALUES (:userId, :title, :updateFrequency, NOW())`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdatePrayerList updates the title and updateFrequency on a list
func UpdatePrayerList(input *PrayerList) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := Config.DbConn.NamedExec(`UPDATE PrayerLists SET title = :title, updateFrequency = :updateFrequency WHERE id = :id LIMIT 1`, input)
	return err
}

// DeletePrayerList deletes a prayer list and all attached request links
func DeletePrayerList(listID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM PrayerLists WHERE id = ?", listID)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM PrayerRequestPrayerListLinks WHERE listId = ?", listID)
	if err != nil {
		return err
	}
	return nil
}

// GetPrayerList gets a prayer list and attached requests
func GetPrayerList(listID int64) (PrayerList, error) {
	list := PrayerList{}
	defer list.processForAPI()
	err := Config.DbConn.Get(&list, "SELECT * FROM PrayerLists WHERE id = ?", listID)
	list.PrayerRequests, _ = GetPrayerRequestsOnPrayerList(listID)
	return list, err
}

// GetPrayerListsForUser gets the liss of lists for the user, optionally sorted
func GetPrayerListsForUser(userID int64, sortField string) ([]PrayerList, error) {
	lists := []PrayerList{}
	var err error
	sortField = strings.ToLower(sortField)
	if sortField == "title" {
		err = Config.DbConn.Select(&lists, "SELECT * FROM PrayerLists WHERE userId = ? ORDER BY title", userID)
	} else {
		err = Config.DbConn.Select(&lists, "SELECT * FROM PrayerLists WHERE userId = ? ORDER BY created", userID)
	}
	return lists, err
}

// GetPrayerRequestsOnPrayerList gets all of the requests on a list
func GetPrayerRequestsOnPrayerList(listID int64) ([]PrayerRequest, error) {
	requests := []PrayerRequest{}
	query := `SELECT pr.*, l.added, (SELECT COUNT(*) FROM Prayers p WHERE p.prayerRequestId = pr.id) AS prayerCount 
	FROM PrayerRequests pr, PrayerRequestPrayerListLinks l 
	WHERE l.listId = ? AND l.prayerRequestId = pr.id ORDER BY l.added`

	err := Config.DbConn.Select(&requests, query, listID)
	for i := range requests {
		requests[i].processForAPI()
	}
	return requests, err
}

// AddRequestToPrayerList adds a request to a list
func AddRequestToPrayerList(requestID, listID int64) error {
	_, err := Config.DbConn.Exec("INSERT INTO PrayerRequestPrayerListLinks (prayerRequestId, listId, added) VALUES (?,?, NOW()) ON DUPLICATE KEY UPDATE prayerRequestId = prayerRequestId", requestID, listID)
	return err
}

// RemoveRequestFromPrayerList removes a request from a list
func RemoveRequestFromPrayerList(requestID, listID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM PrayerRequestPrayerListLinks WHERE prayerRequestId = ? AND listId = ?", requestID, listID)
	return err
}

func (u *PrayerList) processForDB() {
	if u.Created == "" {
		u.Created = "1970-01-01T14:05:06Z"
	} else {
		parsed, err := time.Parse("2006-01-02T15:04:05Z", u.Created)
		if err != nil {
			// perhaps it was already db time
			parsed, err = time.Parse("2006-01-02 15:04:05", u.Created)
			if err != nil {
				// last try; no time?
				parsed, err = time.Parse("2006-01-02", u.Created)
				if err != nil {
					// screw it
					parsed, _ = time.Parse("2006-01-02", "1970-01-01")
				}
			}
		}
		u.Created = parsed.Format("2006-01-02")
	}

	if u.UpdateFrequency == "" {
		u.UpdateFrequency = PrayerListUpdateFrequencyNever
	}
}

func (u *PrayerList) processForAPI() {
	if u.Created == "1970-01-01 00:00:00" {
		u.Created = ""
	} else {
		u.Created, _ = ParseTimeToDate(u.Created)
	}
}
