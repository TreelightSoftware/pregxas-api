package api

import (
	"time"
)

// Community is a group or organization that ties users together
type Community struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	ShortCode string `json:"shortCode" db:"shortCode"`
	Created   string `json:"created" db:"created"`
	// UserSignupStatus is a setting that sets the default status of users who sign up
	UserSignupStatus    string `json:"userSignupStatus" db:"userSignupStatus"`
	Plan                string `json:"plan" db:"plan"`
	PlanPaidThrough     string `json:"planPaidThrough" db:"planPaidThrough"`
	PlanDiscountPercent int64  `json:"planDiscountPercent" db:"planDiscountPercent"`
	StripeChargeToken   string `json:"stripeChargeToken" db:"stripeChargeToken"`
	// UserStatus is only populated in queries in which a user is joined or invited to a community
	UserStatus string `json:"userStatus" db:"userStatus"`
	// UserRole is only populated in queries in which a user is joined or invited to a community
	UserRole string `json:"userRole" db:"userRole"`
}

const (
	// CommunityUserSignupStatusNone indicates users cannot signup for the community, even with a short code
	CommunityUserSignupStatusNone = "none"

	// CommunityUserSignupStatusApproval indicates users can request to join, but it is private by default and needs approval
	CommunityUserSignupStatusApproval = "approval_required"

	// CommunityUserSignupStatusAccept indicates users can signup and will be accepted automatically
	CommunityUserSignupStatusAccept = "auto_accept"

	// CommunityUserLinkStatusInvited indicates a user has been invited to join a community
	CommunityUserLinkStatusInvited = "invited"

	// CommunityUserLinkStatusRequested indicates a user has requested to join a community
	CommunityUserLinkStatusRequested = "requested"

	// CommunityUserLinkStatusAccepted indicates a user is a member of a community
	CommunityUserLinkStatusAccepted = "accepted"

	// CommunityUserLinkStatusDeclined indicates a user has been denied access to a community
	CommunityUserLinkStatusDeclined = "declined"
)

// GetCommunityByShortCode gets a community by its short code
func GetCommunityByShortCode(shortCode string) (*Community, error) {
	found := &Community{}
	err := Config.DbConn.Get(found, "SELECT c.* FROM Communities c WHERE c.shortCode = ?", shortCode)
	found.processForAPI()
	return found, err
}

// GetCommunityByID gets a community by its id
func GetCommunityByID(id int64) (*Community, error) {
	found := &Community{}
	err := Config.DbConn.Get(found, "SELECT c.* FROM Communities c WHERE c.id = ? LIMIT 1", id)
	found.processForAPI()
	return found, err
}

// CreateCommunity creates a new community on the platform
func CreateCommunity(input *Community) error {
	input.processForDB()
	defer input.processForAPI()
	result, err := Config.DbConn.NamedExec(`INSERT INTO Communities (name, shortCode, created, userSignupStatus)
		VALUES (:name, :shortCode, NOW(), :userSignupStatus)`, input)
	if err != nil {
		return err
	}
	input.ID, _ = result.LastInsertId()
	return nil
}

// UpdateCommunity attempts to update a single community
func UpdateCommunity(input *Community) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := Config.DbConn.NamedExec(`UPDATE Communities SET name = :name, shortCode = :shortCode, userSignupStatus = :userSignupStatus WHERE id = :id`, input)
	return err
}

// DeleteCommunity deletes a community and all links
func DeleteCommunity(id int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM Communities WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM CommunityUserLinks WHERE communityId = ?", id)
	return err
}

// CreateCommunityUserLink creates a new link between a user and a community
func CreateCommunityUserLink(communityID, userID int64, role, status string) error {
	if role == "" {
		role = "member"
	}
	_, err := Config.DbConn.Exec("INSERT INTO CommunityUserLinks (communityId, userId, role, status) VALUES (?, ?, ?, ?)", communityID, userID, role, status)
	return err
}

// UpdateCommunityUserLink updates the status of a link
func UpdateCommunityUserLink(communityID, userID int64, status string) error {
	_, err := Config.DbConn.Exec("UPDATE CommunityUserLinks SET status = ? WHERE communityId = ? AND userId = ?", status, communityID, userID)
	return err
}

// DeleteCommunityUserLink completely deletes a link between a user and a community and should only be used by the system
func DeleteCommunityUserLink(communityID, userID int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM CommunityUserLinks WHERE communityId = ? AND userId = ?", communityID, userID)
	return err
}

// GetCountOfUsersInCommunity gets the count of users in a community
func GetCountOfUsersInCommunity(communityID int64) (int64, error) {
	count := struct {
		Count int64 `db:"count"`
	}{}
	err := Config.DbConn.Get(&count, `SELECT COUNT(*) AS count FROM CommunityUserLinks cul WHERE cul.communityId = ? AND cul.status = ?`,
		communityID, CommunityUserLinkStatusAccepted)
	return count.Count, err
}

// GetCommunitiesForUser gets all of the communities for a user as well as their status
func GetCommunitiesForUser(userID int64) ([]Community, error) {
	comms := []Community{}
	err := Config.DbConn.Select(&comms, `SELECT c.*, cul.status AS userStatus, cul.role as userRole 
		FROM Communities c, CommunityUserLinks cul WHERE cul.userId = ? AND cul.communityId = c.id ORDER BY c.name`, userID)
	for i := range comms {
		comms[i].processForAPI()
	}
	return comms, err
}

// GetUsersInCommunity gets all of the users in a community
func GetUsersInCommunity(communityID int64) ([]User, error) {
	users := []User{}
	err := Config.DbConn.Select(&users, `SELECT u.*, cul.status AS communityStatus 
		FROM Users u, CommunityUserLinks cul
		WHERE cul.communityId = ? AND cul.userId = u.id ORDER BY u.lastName, u.firstName`, communityID)
	for i := range users {
		users[i].processForAPI()
	}
	return users, err
}

func (input *Community) processForDB() {
	if input.Created == "" {
		input.Created = "1970-01-01"
	} else {
		parsed, err := time.Parse("2006-01-02T15:04:05Z", input.Created)
		if err != nil {
			// perhaps it was already db time
			parsed, err = time.Parse("2006-01-02 15:04:05", input.Created)
			if err != nil {
				// last try; no time?
				parsed, err = time.Parse("2006-01-02", input.Created)
				if err != nil {
					// screw it
					parsed, _ = time.Parse("2006-01-02", "1970-01-01")
				}
			}
		}
		input.Created = parsed.Format("2006-01-02")
	}
	if input.UserSignupStatus == "" {
		input.UserSignupStatus = CommunityUserSignupStatusApproval
	}
}

func (input *Community) processForAPI() {
	if input.UserSignupStatus == "" {
		input.UserSignupStatus = CommunityUserSignupStatusApproval
	}

	if input.Created == "1970-01-01 00:00:00" {
		input.Created = ""
	} else {
		input.Created, _ = ParseTimeToDate(input.Created)
	}
}
