package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestSiteSetupRoute(t *testing.T) {
	ConfigSetup()
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	key := GenerateSiteKey()
	SetupInitialSite(key)

	enc.Encode(map[string]string{})
	code, _, _ := TestAPICall(http.MethodGet, "/admin/site", b, GetSiteInfoRoute, "", "")
	assert.Equal(t, http.StatusBadRequest, code, "empty data")
	code, res, _ := TestAPICall(http.MethodGet, "/admin/site", b, GetSiteInfoRoute, "", key)
	assert.Equal(t, http.StatusOK, code)
	_, body, _ := UnmarshalTestMap(res)
	site := SiteStruct{}
	mapstructure.Decode(body, &site)
	assert.Equal(t, "Pregxas", site.Name)
	assert.Equal(t, "", site.SecretKey)
	assert.Equal(t, "pending_setup", site.Status)

	// now update it
	input := map[string]interface{}{
		"name":        "My Site",
		"description": "My site description",
		"firstName":   "Kevin",
		"lastName":    "Eaton",
		"email":       "kevin@treelightsoftware.com",
		"username":    "kevineaton",
		"password":    "super_secret!!",
	}
	b.Reset()
	enc.Encode(&input)

	code, res, _ = TestAPICall(http.MethodPost, "/admin/site", b, SetupSiteRoute, "", key)
	assert.Equal(t, http.StatusOK, code)

	// try to login with the user
	loginInput := map[string]string{
		"email":    input["email"].(string),
		"password": input["password"].(string),
	}
	b.Reset()
	enc.Encode(&loginInput)
	code, res, _ = TestAPICall(http.MethodPost, "/users/login", b, LoginUserRoute, "", "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	userID, _ := convertTestJSONFloatToInt(body["id"])
	assert.NotZero(t, userID)
	defer DeleteUser(userID)

	assert.Equal(t, input["name"].(string), Site.Name)
	assert.Equal(t, input["description"].(string), Site.Description)
	DeleteSiteForTest()
}
