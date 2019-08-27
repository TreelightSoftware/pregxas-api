package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const (
	// TokenEmailVerify is used to verify a signup email or new email
	TokenEmailVerify = "email"
	// TokenPasswordReset is to verify a forgotten password
	TokenPasswordReset = "password_reset"
)

// Token represents a token, such as for a password or email verification
type Token struct {
	ID        int64  `json:"id" db:"id"`
	Token     string `json:"token" db:"token"`
	TokenType string `json:"tokenType" db:"tokenType"`
	Created   string `json:"created" db:"created"`
	UserID    int64  `json:"userId" db:"userId"`
}

// GenerateToken generates a new token for the user
func GenerateToken(userID int64, tokenType string) (token string, err error) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63n(999999999999)
	str := fmt.Sprintf("%d%d-%d %s", userID, rand.Intn(100000000), r, tokenType)
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	token = hash[0:8]
	_, err = Config.DbConn.Exec("INSERT INTO UserTokens (token, created, tokenType, userId) VALUES (?, NOW(), ?, ?)", token, tokenType, userID)
	return token, err
}

// GetTokenForTest is a helper function for testing
func GetTokenForTest(userID int64, tokenType string) (string, error) {
	token := Token{}
	err := Config.DbConn.Get(&token, "SELECT * FROM UserTokens WHERE userId = ? AND tokenType = ? LIMIT 1", userID, tokenType)
	return token.Token, err
}

// VerifyToken verifies a token. If the token is verified, it is removed from the DB
func VerifyToken(token, tokenType string) (userID int64, valid bool, err error) {
	found := Token{}
	err = Config.DbConn.Get(&found, "SELECT * FROM UserTokens WHERE token = ? AND tokenType = ?", token, tokenType)
	if err != nil {
		return 0, false, err
	}
	userID = found.UserID
	valid = true
	Config.DbConn.Exec("DELETE FROM UserTokens WHERE id = ?", found.ID)
	return
}

// DeleteTokensCreatedBeforeTime deletes tokens created before a specific time
func DeleteTokensCreatedBeforeTime(time string) error {
	_, err := Config.DbConn.Exec("DELETE FROM UserTokens WHERE created < ?", time)
	return err
}

// GenerateRandomPassword generates a random password for a user; the _ at the start indicates that it should be changed if it is decrypted. This should
// really only be used in cases of user imports or automatic creation via a community.
func GenerateRandomPassword(input *User) string {
	randID := rand.NewSource(time.Now().UnixNano()).Int63()
	str := fmt.Sprintf("%s-%s-%d-%s^%s", input.FirstName, time.Now().Format("2006-01-02-15:04:05"), randID, input.LastName, input.Email)
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	token := "_" + hash[0:28]
	return token
}

// GenerateSiteKey generates a new site key for setup
func GenerateSiteKey() string {
	randID := rand.NewSource(time.Now().UnixNano()).Int63()
	str := fmt.Sprintf("%s-%d-^%s", time.Now().Format("2006-01-02-15:04:05"), randID, "pregxas-setup")
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	token := "_" + hash[0:28]
	return token
}

// GenerateShortCode generates a membership request token
func GenerateShortCode(communityID, userID int64) string {
	randID := rand.NewSource(time.Now().UnixNano()).Int63()
	str := fmt.Sprintf("%s-%d-^%s!!%d<>%d", time.Now().Format("2006-01-02-15:04:05"), randID, "pregxas-request", communityID, userID)
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	token := "_" + hash[0:9]
	return token
}
