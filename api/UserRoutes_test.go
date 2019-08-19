package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSignupPaths(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999999)
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	// try to signup with no data
	enc.Encode(map[string]string{})
	code, _, _ := TestAPICall(http.MethodPost, "/users/signup", b, SignupUserRoute, "", "")
	assert.Equal(t, http.StatusBadRequest, code, "empty data")

	// now sign up
	input := map[string]string{
		"firstName": "Kevin",
		"lastName":  "Eaton",
		"email":     fmt.Sprintf("Test-%d@pregxas", randID),
		"username":  fmt.Sprintf("test-%d", randID),
		"password":  fmt.Sprintf("pass-%d", randID),
	}
	b.Reset()
	enc.Encode(&input)
	code, res, _ := TestAPICall(http.MethodPost, "/users/signup", b, SignupUserRoute, "", "")
	require.Equal(t, http.StatusCreated, code)
	_, body, _ := UnmarshalTestMap(res)
	id, err := convertTestJSONFloatToInt(body["id"])
	assert.Nil(t, err)
	defer DeleteUser(id)

	// try to login; should fail since not verified
	loginInput := map[string]string{
		"email":    input["email"],
		"password": input["password"],
	}
	b.Reset()
	enc.Encode(&loginInput)
	code, res, _ = TestAPICall(http.MethodPost, "/users/login", b, LoginUserRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)

	// email isn't sent in tests, so let's grab it from the DB
	token, err := GetTokenForTest(id, "email")
	assert.Nil(t, err)

	verifyInput := map[string]string{
		"token": token,
		"email": input["email"],
	}
	b.Reset()
	enc.Encode(&verifyInput)
	code, res, _ = TestAPICall(http.MethodPost, "/users/signup/verify", b, VerifyEmailAndTokenRoute, "", "")
	assert.Equal(t, http.StatusOK, code)

	b.Reset()
	enc.Encode(&loginInput)
	code, res, _ = TestAPICall(http.MethodPost, "/users/login", b, LoginUserRoute, "", "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)

	user := User{}
	err = mapstructure.Decode(body, &user)
	assert.Nil(t, err)
	assert.Equal(t, "", user.Password)
	assert.NotEqual(t, "", user.JWT)
	assert.Equal(t, input["firstName"], user.FirstName)
	assert.Equal(t, input["lastName"], user.LastName)
	assert.Equal(t, strings.ToLower(input["email"]), user.Email)

	// get the profile
	code, res, _ = TestAPICall(http.MethodGet, "/me", b, GetMyProfileRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)

	profile := User{}
	err = mapstructure.Decode(body, &profile)
	assert.Equal(t, "", profile.Password)
	assert.Equal(t, "", profile.JWT)
	assert.Equal(t, input["firstName"], profile.FirstName)
	assert.Equal(t, input["lastName"], profile.LastName)
	assert.Equal(t, strings.ToLower(input["email"]), profile.Email)

	// update the profile and try again
	profileUpdateInput := map[string]string{
		"firstName": "Updated First",
		"lastName": "Updated Last",
		"username": "updated-user",
	}
	b.Reset()
	enc.Encode(&profileUpdateInput)
	code, res, _ = TestAPICall(http.MethodPatch, "/me", b, GetMyProfileRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, "/me", b, GetMyProfileRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)

	updated := User{}
	err = mapstructure.Decode(body, &updated)
	assert.Equal(t, "", updated.Password)
	assert.Equal(t, "", updated.JWT)
	assert.Equal(t, profileUpdateInput["firstName"], updated.FirstName)
	assert.Equal(t, profileUpdateInput["lastName"], updated.LastName)
	assert.Equal(t, profileUpdateInput["username"], updated.Username)
	assert.Equal(t, strings.ToLower(input["email"]), updated.Email)
}

func TestPasswordResetRoutes(t *testing.T){
	ConfigSetup()
	randID := rand.Int63n(99999999)
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	user := User{}
	err := CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user)

	resetInput := map[string]string{
		"email": user.Email,
	}
	b.Reset()
	enc.Encode(&resetInput)
	code, _, _ := TestAPICall(http.MethodPost, "/users/login/reset", b, ResetPasswordStartRoute, "", "")
	require.Equal(t, http.StatusOK, code)

	// find the code
	token, err := GetTokenForTest(user.ID, TokenPasswordReset)
	require.Nil(t, err)

	verifyInput := map[string]string{
		"email": user.Email,
		"password": fmt.Sprintf("pass-%d", randID),
		"token": token,
	}
	b.Reset()
	enc.Encode(&verifyInput)
	code, _, _ = TestAPICall(http.MethodPost, "/users/login/reset/verify", b, ResetPasswordVerifyRoute, "", "")
	assert.Equal(t, http.StatusOK, code)

	// login with the new password
	loginInput := map[string]string{
		"email": user.Email,
		"password": verifyInput["password"],
	}
	b.Reset()
	enc.Encode(&loginInput)
	code, res, _ := TestAPICall(http.MethodPost, "/users/login", b, LoginUserRoute, "", "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ := UnmarshalTestMap(res)

	loggedIn := User{}
	err = mapstructure.Decode(body, &loggedIn)
	assert.Nil(t, err)
	assert.Equal(t, "", loggedIn.Password)
	assert.NotEqual(t, "", loggedIn.JWT)
	assert.Equal(t, user.FirstName, loggedIn.FirstName)
	assert.Equal(t, user.LastName, loggedIn.LastName)
	assert.Equal(t, user.Email, loggedIn.Email)

}