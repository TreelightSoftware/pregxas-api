package api

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"golang.org/x/crypto/bcrypt"
)

// User is a user on the platform, including "users", "students", and "admins"
type User struct {
	ID           int64  `json:"id" db:"id"`
	FirstName    string `json:"firstName" db:"firstName"`
	LastName     string `json:"lastName" db:"lastName"`
	Email        string `json:"email" db:"email"`
	Created      string `json:"created" db:"created"`
	Password     string `json:"password,omitempty" db:"password"`
	Status       string `json:"status" db:"status"`
	Username     string `json:"username" db:"username"`
	Updated      string `json:"updated" db:"updated"`
	LastLogin    string `json:"lastLogin" db:"lastLogin"`
	JWT          string `json:"-" db:"-"` // needed for tests at this time, no longer sent to the user
	PlatformRole string `json:"platformRole" db:"platformRole"`
	// CommunityStatus represents the user's status in a given community; used in queries with joins
	CommunityStatus string `json:"communityStatus,omitempty" db:"communityStatus"`

	// the tokens are used for session and state; to be closer to the eventual addition of oAuth2, notice the use of underscores instead of camel case
	AccessToken  string `json:"access_token,omitempty" db:"-"`
	RefreshToken string `json:"refresh_token,omitempty" db:"-"`
	ExpiresIn    int64  `json:"expires_in,omitempty" db:"-"`
	ExpiresAt    string `json:"expires_at,omitempty" db:"-"`
}

const (
	// UserStatusPending represents a user status of pending and needing verification
	UserStatusPending = "pending"
	// UserStatusVerified represents an active user
	UserStatusVerified = "verified"

	// AccessTokenExpiresSeconds is how long an access token is valid for
	AccessTokenExpiresSeconds = 60 * 60 // 1 hour to start
)

// CreateUser creates a new user in the system
func CreateUser(input *User) error {
	input.processForDB()
	defer input.processForAPI()
	query := `INSERT INTO Users (firstName, lastName, email, created, password, status, username, updated, lastLogin, platformRole) 
	VALUES (:firstName, :lastName, :email, NOW(), :password, :status, :username, NOW(), :lastLogin, :platformRole)
	`
	res, err := Config.DbConn.NamedExec(query, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// CreateTestUser creates a user for testing
func CreateTestUser(input *User) error {
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63n(9999999999)

	//essentially, we let them pass in whatever but give defaults for anything not sent in

	if input.FirstName == "" {
		input.FirstName = fmt.Sprintf("First-%d", r)
	}
	if input.LastName == "" {
		input.LastName = fmt.Sprintf("Last-%d", r)
	}
	if input.Email == "" {
		input.Email = fmt.Sprintf("user-%d@pregxas.com", r)
	}
	if input.Username == "" {
		input.Username = fmt.Sprintf("user-%d", r)
	}

	if input.Password == "" {
		input.Password, _ = encrypt(fmt.Sprintf("pass-%d", r))
	}
	if input.Status == "" {
		input.Status = "verified"
	}

	if input.PlatformRole == "" {
		input.PlatformRole = "member"
	}

	err := CreateUser(input)
	return err
}

// DeleteUserFromTest deletes a test user; a placeholder in case tests need further teardown
func DeleteUserFromTest(user *User) {
	DeleteUser(user.ID)
}

// UpdateUser updates a user's profile and specific settings
func UpdateUser(input *User) error {
	input.processForDB()
	defer input.processForAPI()

	query := `UPDATE Users SET firstName = :firstName, lastName = :lastName, email = :email, username = :username, password = :password, status = :status, updated = NOW() WHERE id = :id LIMIT 1`
	_, err := Config.DbConn.NamedExec(query, input)
	input.processForAPI()
	return err
}

// GetUserByID gets a user by its ID
func GetUserByID(userID int64) (*User, error) {
	found := &User{}
	err := Config.DbConn.Get(found, "SELECT * FROM Users WHERE id = ? LIMIT 1", userID)
	found.processForAPI()
	return found, err
}

// GetUserByEmail gets a user by email address
func GetUserByEmail(email string) (*User, error) {
	found := &User{}
	err := Config.DbConn.Get(found, "SELECT * FROM Users WHERE email = ?", email)
	found.processForAPI()
	return found, err
}

// GetUserByUsername gets a user by username
func GetUserByUsername(username string) (*User, error) {
	found := &User{}
	err := Config.DbConn.Get(found, "SELECT * FROM Users WHERE username = ?", username)
	found.processForAPI()
	return found, err
}

// DeleteUser deletes a user from the DB
func DeleteUser(userID int64) {
	//we need to delete the user, all tokens, literally everything
	Config.DbConn.Exec("DELETE FROM Users where id = ?", userID)
	Config.DbConn.Exec("DELETE FROM CommunityUserLinks where userId = ?", userID)
	Config.DbConn.Exec("DELETE FROM Prayers where userId = ?", userID)
}

// LoginUser attempts to login a user
func LoginUser(email, plainPassword string) (found *User, err error) {
	found = &User{}
	err = Config.DbConn.Get(found, "SELECT * FROM Users WHERE email = ? LIMIT 1", email)
	if err != nil || found == nil || found.Email == "" {
		return
	}
	passwordCorrect := checkEncrypted(plainPassword, found.Password)
	if !passwordCorrect {
		err = errors.New("incorrect password")
		return
	}
	//we are good, jwt it, update last logged in and move on
	Config.DbConn.Exec("UPDATE Users SET lastLogin = NOW() where id = ?", found.ID)
	found.LastLogin = time.Now().Format("2006-01-02 15:04:05")
	found.processForAPI()
	return
}

// processForDB ensures data consistency
func (u *User) processForDB() {

	if u.Created == "" {
		u.Created = "1970-01-01 00:00:00"
	} else {
		parsed, err := ParseTime(u.Created)
		if err != nil {
			parsed = time.Now()
		}
		u.Created = parsed.Format("2006-01-02 15:04:05")
	}

	if u.Updated == "" {
		u.Updated = "1970-01-01 00:00:00"
	} else {
		parsed, err := ParseTime(u.Updated)
		if err != nil {
			parsed = time.Now()
		}
		u.Updated = parsed.Format("2006-01-02 15:04:05")
	}

	if u.LastLogin == "" {
		u.LastLogin = "1970-01-01 00:00:00"
	} else {
		parsed, err := ParseTime(u.LastLogin)
		if err != nil {
			parsed = time.Now()
		}
		u.LastLogin = parsed.Format("2006-01-02 15:04:05")
	}

	if u.Status == "" {
		u.Status = "pending"
	}

	// check if we need to change the password
	if u.Password != "" && !strings.HasPrefix(u.Password, "$2a$") {
		// we have a plaintext password, so hash it
		hashed, err := encrypt(u.Password)
		if err == nil {
			u.Password = hashed
		}
	}

	if u.PlatformRole == "" {
		u.PlatformRole = "member"
	}
}

// processForAPI ensures data consistency and creates the JWT
func (u *User) processForAPI() {
	if u == nil {
		return
	}

	if u.Created == "1970-01-01 00:00:00" {
		u.Created = ""
	} else {
		u.Created, _ = ParseTimeToISO(u.Created)
	}

	if u.Updated == "1970-01-01 00:00:00" {
		u.Updated = ""
	} else {
		u.Updated, _ = ParseTimeToISO(u.Updated)
	}

	if u.LastLogin == "1970-01-01 00:00:00" {
		u.LastLogin = ""
	} else {
		u.LastLogin, _ = ParseTimeToISO(u.LastLogin)
	}

	if u.Status == "" {
		u.Status = "pending"
	}
}

func (u *User) clean() {
	u.Password = ""
	u.JWT = ""
}

// JWTUser is a user decrypted from a JWT token
type JWTUser struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Expires      string   `json:"expires"`
	ExpiresIn    int64    `json:"expires_in"`
	PlatformRole string   `json:"platformRole"`
	Scopes       []string `json:"scopes"`
}

type jwtClaims struct {
	User JWTUser `json:"user"`
	jwt.StandardClaims
}

// createJwt creates a new jwt for a user
func createJwt(payload *User) (string, error) {
	if payload.PlatformRole == "" {
		payload.PlatformRole = "member"
	}
	jwtu := JWTUser{
		ID:           payload.ID,
		Email:        payload.Email,
		Expires:      time.Now().Add(time.Second * AccessTokenExpiresSeconds).Format("2006-01-02T15:04:05Z"),
		ExpiresIn:    AccessTokenExpiresSeconds,
		PlatformRole: payload.PlatformRole,
		Scopes:       []string{}, // not currently used
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": jwtu,
	})
	tokenString, err := token.SignedString([]byte(Config.JWTSigningString))

	return tokenString, err
}

func parseJwt(jwtString string) (JWTUser, error) {
	token, err := jwt.ParseWithClaims(jwtString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(Config.JWTSigningString), nil
	})
	if err != nil {
		return JWTUser{}, errors.New("Could not parse jwt")
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		u := claims.User
		// TODO: check expiration
		return u, nil
	}
	return JWTUser{}, errors.New("Could not parse jwt")
}

func encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkEncrypted(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
