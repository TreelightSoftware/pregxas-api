package api

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	ConfigSetup()
	input := User{}
	err := CreateTestUser(&input)
	assert.Nil(t, err)

	jwt, err := createJwt(&input)
	assert.Nil(t, err)
	assert.NotEqual(t, "", jwt)

	parsed, err := parseJwt(jwt)
	assert.Nil(t, err)
	assert.Equal(t, input.ID, parsed.ID)
	assert.Equal(t, input.Email, parsed.Email)

	DeleteUser(input.ID)
}

func TestUserCRUD(t *testing.T) {
	ConfigSetup()
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63n(999999999)
	user := User{
		FirstName: "Kevin",
		LastName:  "Eaton",
		Email:     fmt.Sprintf("test-%d@pregxas.com", r),
		Password:  "password",
	}
	err := CreateUser(&user)
	assert.Nil(t, err)
	assert.NotZero(t, user.ID)
	defer DeleteUser(user.ID)

	found, err := GetUserByID(user.ID)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.FirstName, found.FirstName)
	assert.Equal(t, user.LastName, found.LastName)
	assert.Equal(t, user.Email, found.Email)

	// try to login
	login, err := LoginUser(user.Email, "password")
	assert.Nil(t, err)
	assert.NotNil(t, login)
	assert.Equal(t, login.ID, user.ID)

	foundByEmail, err := GetUserByEmail(user.Email)
	assert.Nil(t, err)
	assert.Equal(t, user.ID, foundByEmail.ID)

	foundByUsername, err := GetUserByUsername(user.Username)
	assert.Nil(t, err)
	assert.Equal(t, user.ID, foundByUsername.ID)

	// update
	user.FirstName = "Vicki"
	user.LastName = "Alm"
	user.Email = fmt.Sprintf("vicki+%d@pregxas.com", r)
	user.Status = "verified"
	user.Password = "reset"
	err = UpdateUser(&user)
	assert.Nil(t, err)
	login, err = LoginUser(user.Email, "password")
	assert.NotNil(t, err)
	login, err = LoginUser(user.Email, "reset")
	assert.Nil(t, err)

	found, err = GetUserByID(user.ID)
	assert.Equal(t, user.FirstName, found.FirstName)
	assert.Equal(t, user.LastName, found.LastName)
	assert.Equal(t, user.Email, found.Email)
}
