package api

import "strings"

// PrayerRequestTag is a tag applied to a prayer request
type PrayerRequestTag struct {
	ID  int64  `json:"id" db:"id"`
	Tag string `json:"tag" db:"tag"`
}

// PrayerRequestTagLinks is a prayer request and tag link combination
type PrayerRequestTagLinks struct {
	TagID           int64 `json:"tagId" db:"tagId"`
	PrayerRequestID int64 `json:"prayerRequestId" db:"prayerRequestId"`
}

// AddTagToPrayerRequest adds a tag to a request; if the tag doesn't exist it will be created
func AddTagToPrayerRequest(prayerRequestID int64, tag string) (PrayerRequestTag, error) {
	tag = strings.ToLower(tag)
	// first, find out if that tag exists already; if it does, link it to the existing
	// if not, create it first
	found, err := GetTagIDByTag(tag)
	if err != nil || found.ID == 0 {
		// we need to create the tag
		res, err := Config.DbConn.Exec("INSERT INTO PrayerRequestTags (tag) VALUES (?)", tag)
		if err != nil {
			return found, err
		}
		found.ID, _ = res.LastInsertId()
		found.Tag = tag
	}
	// link it
	_, err = Config.DbConn.Exec("INSERT INTO PrayerRequestTagLinks (prayerRequestId, tagId) VALUES (?, ?)", prayerRequestID, found.ID)
	return found, err
}

// GetTagIDByTag gets a tag by its name
func GetTagIDByTag(tag string) (PrayerRequestTag, error) {
	tag = strings.ToLower(tag)
	existingTag := PrayerRequestTag{}
	err := Config.DbConn.Get(&existingTag, "SELECT * FROM PrayerRequestTags WHERE tag = ?", tag)
	return existingTag, err
}

// GetTagsOnRequest gets the tags on a request
func GetTagsOnRequest(prayerRequestID int64) ([]PrayerRequestTag, error) {
	tags := []PrayerRequestTag{}
	err := Config.DbConn.Select(&tags, "SELECT t.* FROM PrayerRequestTagLinks l, PrayerRequestTags t WHERE l.prayerRequestId = ? AND l.tagId = t.id ORDER BY t.tag", prayerRequestID)
	return tags, err
}

// RemoveTagFromRequest unlinks a tag and a request; a cron will clean up any orphaned tags we don't want to keep around
func RemoveTagFromRequest(prayerRequestID, tagID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM PrayerRequestTagLinks WHERE prayerRequestId = ? AND tagId = ? LIMIT 1", prayerRequestID, tagID)
	return err
}

// DeleteTag deletes a tag from the database and should only be used by admins or tests
func DeleteTag(tagID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM PrayerRequestTagLinks WHERE tagId = ?", tagID)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM PrayerRequestTags WHERE id = ?", tagID)
	if err != nil {
		return err
	}
	return nil
}
