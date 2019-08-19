package api

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokensCRD(t *testing.T) {
	ConfigSetup()
	r := rand.Int63n(99999999)
	token, err := GenerateToken(r, "email")
	assert.Nil(t, err)
	userID, valid, err := VerifyToken("moo", "email")
	assert.Zero(t, userID)
	assert.False(t, valid)
	assert.NotNil(t, err)

	userID, valid, err = VerifyToken(token, "email")
	assert.Equal(t, r, userID)
	assert.True(t, valid)
	assert.Nil(t, err)

	userID, valid, err = VerifyToken(token, "email")
	assert.Zero(t, userID)
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestTokensCron(t *testing.T) {
	ConfigSetup()
	r := rand.Int63n(99999999)
	fiveMinutes := time.Now().Add(-5 * time.Minute).Format("2006-01-02 15:04:05")
	fiveToken := fmt.Sprintf("test-5-%d", r)
	tenMinutes := time.Now().Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
	twentyMinutes := time.Now().Add(-20 * time.Minute).Format("2006-01-02 15:04:05")
	twentyToken := fmt.Sprintf("test-20-%d", r)
	// we will directly insert two
	res, err := Config.DbConn.Exec("INSERT INTO UserTokens (token, createdOn, tokenType, userId) VALUES (?, ?, ?, ?)", fiveToken, fiveMinutes, "email", r)
	assert.Nil(t, err)
	token1ID, _ := res.LastInsertId()
	// we will directly insert two
	res, err = Config.DbConn.Exec("INSERT INTO UserTokens (token, createdOn, tokenType, userId) VALUES (?, ?, ?, ?)", twentyToken, twentyMinutes, "email", r)
	assert.Nil(t, err)

	err = DeleteTokensCreatedBeforeTime(tenMinutes)
	assert.Nil(t, err)
	id, valid, err := VerifyToken(twentyToken, "email")
	assert.Zero(t, id)
	assert.False(t, valid)
	assert.NotNil(t, err)
	id, valid, err = VerifyToken(fiveToken, "email")
	assert.NotZero(t, id)
	assert.True(t, valid)
	assert.Nil(t, err)

	// delete token 2
	Config.DbConn.Exec("DELETE FROM UserTokens WHERE id = ?", token1ID)

}

func TestRandomPassword(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999)
	user := User{
		FirstName: fmt.Sprintf("%d-first", randID),
		LastName:  fmt.Sprintf("%d-last", randID),
		Email:     fmt.Sprintf("%d@pregxas.com", randID),
	}
	random1 := GenerateRandomPassword(&user)
	assert.NotEqual(t, "", random1)

	user.FirstName = user.FirstName + "up"
	user.LastName = user.LastName + "up"
	user.Email = user.Email + "up"
	random2 := GenerateRandomPassword(&user)
	assert.NotEqual(t, "", random2)
	assert.NotEqual(t, random1, random2)
}
