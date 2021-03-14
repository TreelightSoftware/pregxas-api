package api

import (
	"fmt"
	"strings"
	"time"
)

// Community is a group or organization that ties users together
type Community struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	ShortCode   string `json:"shortCode,omitempty" db:"shortCode"` // this is a lookup for a specific community
	JoinCode    string `json:"joinCode,omitempty" db:"joinCode"`   // this is a code used to join a community
	Created     string `json:"created" db:"created"`
	Privacy     string `json:"privacy" db:"privacy"`
	// UserSignupStatus is a setting that sets the default status of users who sign up
	UserSignupStatus     string `json:"userSignupStatus,omitempty" db:"userSignupStatus"`
	Plan                 string `json:"plan" db:"plan"`
	PlanPaidThrough      string `json:"planPaidThrough,omitempty" db:"planPaidThrough"`
	PlanDiscountPercent  int64  `json:"planDiscountPercent,omitempty" db:"planDiscountPercent"`
	StripeSubscriptionID string `json:"stripeSubscriptionId,omitempty" db:"stripeSubscriptionId"`
	// UserStatus is only populated in queries in which a user is joined or invited to a community
	UserStatus string `json:"userStatus,omitempty" db:"userStatus"`
	// UserRole is only populated in queries in which a user is joined or invited to a community
	UserRole string `json:"userRole,omitempty" db:"userRole"`

	MemberCount  int64 `json:"memberCount" db:"memberCount"`
	RequestCount int64 `json:"requestCount" db:"requestCount"`
}

// CommunityUserLink is a link between a user and a community
type CommunityUserLink struct {
	UserID      int64  `json:"userId" db:"userId"`
	CommunityID int64  `json:"communityId" db:"communityId"`
	Role        string `json:"role" db:"role"`
	Status      string `json:"status" db:"status"`
	ShortCode   string `json:"shortCode" db:"shortCode"`
	FirstName   string `json:"firstName" db:"firstName"`
	LastName    string `json:"lastName" db:"lastName"`
	Email       string `json:"email" db:"email"`
	Username    string `json:"username" db:"username"`
}

// CommunityPlan is the details for the plans
type CommunityPlan struct {
	AllowedUsers          int64 `json:"allowedUsers"`
	AllowedActiveRequests int64 `json:"allowedActiveRequests"`
	MonthlyPrice          int64 `json:"monthlyPrice"`
}

var plans = map[string]CommunityPlan{
	CommunityPlanFree: {
		AllowedUsers:          50,
		AllowedActiveRequests: 50,
		MonthlyPrice:          0,
	},
	CommunityPlanBasic: {
		AllowedUsers:          200,
		AllowedActiveRequests: 500,
		MonthlyPrice:          499,
	},
	CommunityPlanPro: {
		AllowedUsers:          2000,
		AllowedActiveRequests: 4000,
		MonthlyPrice:          999,
	},
}

const (
	// CommunityUserSignupStatusNone indicates users cannot signup for the community, even with a short code
	CommunityUserSignupStatusNone = "none"

	// CommunityUserSignupStatusJoinCode indicates that a user can join the community with a join code
	CommunityUserSignupStatusJoinCode = "join_code"

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

	// CommunityPrivacyPrivate is a private community not listed in the public directory
	CommunityPrivacyPrivate = "private"

	// CommunityPrivacyPublic means the community is listed in the public directory
	CommunityPrivacyPublic = "public"

	// CommunityPlanFree is a basic free community
	CommunityPlanFree = "free"

	// CommunityPlanBasic represents the basic plan, allowing more users and requests
	CommunityPlanBasic = "basic"

	// CommunityPlanPro represents the pro plan, which allows a lot more requests and users
	CommunityPlanPro = "pro"
)

// GetCommunityByShortCode gets a community by its short code
func GetCommunityByShortCode(shortCode string) (*Community, error) {
	found := &Community{}
	err := Config.DbConn.Get(found, `SELECT c.*,
	(SELECT COUNT(*) FROM CommunityUserLinks cul WHERE cul.communityId = c.id AND cul.status = 'accepted') AS memberCount,
	(SELECT COUNT(*) FROM PrayerRequestCommunityLinks prcl WHERE prcl.communityId = c.id) AS requestCount 
	FROM Communities c WHERE c.shortCode = ?`, shortCode)
	found.processForAPI()
	return found, err
}

// GetCommunityByName is used when a community is being created or updated to make sure that the
// community doesn't already exist (to avoid confusion)
func GetCommunityByName(name string) (*Community, error) {
	found := &Community{}
	err := Config.DbConn.Get(found, `SELECT c.*,
	(SELECT COUNT(*) FROM CommunityUserLinks cul WHERE cul.communityId = c.id AND cul.status = 'accepted') AS memberCount,
	(SELECT COUNT(*) FROM PrayerRequestCommunityLinks prcl WHERE prcl.communityId = c.id) AS requestCount 
	FROM Communities c WHERE c.name = ?`, name)
	found.processForAPI()
	return found, err
}

// GetCommunityByID gets a community by its id
func GetCommunityByID(id int64) (*Community, error) {
	found := &Community{}
	err := Config.DbConn.Get(found, `SELECT c.* ,
	(SELECT COUNT(*) FROM CommunityUserLinks cul WHERE cul.communityId = c.id AND cul.status = 'accepted') AS memberCount,
	(SELECT COUNT(*) FROM PrayerRequestCommunityLinks prcl WHERE prcl.communityId = c.id) AS requestCount
	FROM Communities c WHERE c.id = ? LIMIT 1`, id)
	found.processForAPI()
	return found, err
}

// CreateCommunity creates a new community on the platform
func CreateCommunity(input *Community) error {
	input.processForDB()
	defer input.processForAPI()
	result, err := Config.DbConn.NamedExec(`INSERT INTO Communities (name, description, shortCode, joinCode, created, userSignupStatus, privacy)
		VALUES (:name, :description, :shortCode, :joinCode, NOW(), :userSignupStatus, :privacy)`, input)
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
	_, err := Config.DbConn.NamedExec(`UPDATE Communities SET name = :name, description = :description, shortCode = :shortCode, joinCode = :joinCode, userSignupStatus = :userSignupStatus, privacy = :privacy WHERE id = :id`, input)
	return err
}

// DeleteCommunity deletes a community and all links
func DeleteCommunity(id int64) error {
	_, err := Config.DbConn.Exec("DELETE FROM Communities WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM CommunityUserLinks WHERE communityId = ?", id)
	if err != nil {
		return err
	}
	_, err = Config.DbConn.Exec("DELETE FROM PrayerRequestCommunityLinks WHERE communityId = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// CreateCommunityUserLink creates a new link between a user and a community
func CreateCommunityUserLink(communityID, userID int64, role, status, shortCode string) error {
	if role == "" {
		role = "member"
	}
	_, err := Config.DbConn.Exec("INSERT INTO CommunityUserLinks (communityId, userId, role, status, shortCode) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE userId = userId", communityID, userID, role, status, shortCode)
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

// GetCommunityUserLinks gets the links for a community optionally filtered by status
func GetCommunityUserLinks(communityID int64, status string) ([]CommunityUserLink, error) {
	links := []CommunityUserLink{}
	var err error
	if status != "invited" && status != "requested" && status != "accepted" && status != "declined" {
		err = Config.DbConn.Select(&links, `SELECT cul.*, u.firstName, u.lastName, u.email, u.username FROM CommunityUserLinks cul, Users u 
			WHERE cul.communityId = ? AND cul.userId = u.id ORDER BY u.username`, communityID)
	} else {
		err = Config.DbConn.Select(&links, `SELECT cul.*, u.firstName, u.lastName, u.email, u.username FROM CommunityUserLinks cul, Users u 
			WHERE cul.communityId = ? AND cul.status = ? AND cul.userId = u.id ORDER BY u.username`, communityID, status)
	}
	return links, err
}

// GetCommunityUserLink gets the individual link, used for processing
func GetCommunityUserLink(communityID, userID int64) (CommunityUserLink, error) {
	link := CommunityUserLink{}
	err := Config.DbConn.Get(&link, `SELECT cul.*, u.firstName, u.lastName, u.email, u.username FROM CommunityUserLinks cul, Users u 
		WHERE cul.communityId = ? AND cul.userId = ? AND cul.userId = u.id ORDER BY u.username`, communityID, userID)

	return link, err
}

// GetUserRoleForCommunity gets the user's role for the community
func GetUserRoleForCommunity(communityID, userID int64) (string, error) {
	role := struct {
		Role string `db:"role"`
	}{}
	err := Config.DbConn.Get(&role, "SELECT role FROM CommunityUserLinks WHERE userId = ? AND communityId = ? AND status = 'accepted'", userID, communityID)
	if err != nil {
		return "", err
	}
	return role.Role, nil
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
	err := Config.DbConn.Select(&comms, `SELECT c.*, cul.status AS userStatus, cul.role as userRole,
	(SELECT COUNT(*) FROM CommunityUserLinks cul WHERE cul.communityId = c.id AND cul.status = 'accepted') AS memberCount,
	(SELECT COUNT(*) FROM PrayerRequestCommunityLinks prcl WHERE prcl.communityId = c.id) AS requestCount
	FROM Communities c, CommunityUserLinks cul WHERE cul.userId = ? AND cul.communityId = c.id ORDER BY c.name`, userID)
	for i := range comms {
		comms[i].processForAPI()
	}
	return comms, err
}

// GetPublicCommunities gets the public communities, sorted by either name or created
func GetPublicCommunities(sortField, sortDir string, count, offset int) ([]Community, error) {
	comms := []Community{}

	// generally, string interpolation on SQL is Very Bad, but this is white listed so there is no
	// security concerns
	sortDir = strings.ToUpper(sortDir)
	if sortDir != "DESC" {
		sortDir = "ASC"
	}
	sortField = strings.ToLower(sortField)
	if sortField != "name" && sortField != "created" {
		sortField = "name"
	}
	err := Config.DbConn.Select(&comms, fmt.Sprintf(`
	SELECT c.*, 
	(SELECT COUNT(*) FROM CommunityUserLinks cul WHERE cul.communityId = c.id AND cul.status = 'accepted') AS memberCount,
	(SELECT COUNT(*) FROM PrayerRequestCommunityLinks prcl WHERE prcl.communityId = c.id) AS requestCount
	FROM Communities c
	WHERE c.privacy = 'public'
	ORDER BY c.%s %s LIMIT ?,?`, sortField, sortDir), offset, count)

	for i := range comms {
		comms[i].processForAPI()
		comms[i].clean()
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
	if input.Privacy == "" {
		input.Privacy = "private"
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

	if input.PlanPaidThrough == "1970-01-01" {
		input.PlanPaidThrough = ""
	}
}

func (input *Community) clean() {
	input.StripeSubscriptionID = ""
	input.PlanDiscountPercent = 0
	input.JoinCode = ""
}
