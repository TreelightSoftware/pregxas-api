package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestSiteSetupRoute(t *testing.T) {
	ConfigSetup()
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	randID := rand.Int63n(99999999)
	key := GenerateSiteKey()
	SetupInitialSite(key)

	enc.Encode(map[string]string{})
	code, res, _ := TestAPICall(http.MethodGet, "/admin/site", b, GetSiteInfoRoute, "", key)
	assert.Equal(t, http.StatusOK, code)
	_, body, _ := UnmarshalTestMap(res)
	site := SiteStruct{}
	mapstructure.Decode(body, &site)
	assert.Equal(t, "Pregxas", site.Name)
	assert.Equal(t, "", site.SecretKey)
	assert.Equal(t, "pending_setup", site.Status)

	code, res, _ = TestAPICall(http.MethodPost, "/admin/site", b, SetupSiteRoute, "", "")
	assert.Equal(t, http.StatusBadRequest, code)
	code, res, _ = TestAPICall(http.MethodPost, "/admin/site", b, SetupSiteRoute, "", "moo")
	assert.Equal(t, http.StatusForbidden, code)
	b.Reset()
	enc.Encode(map[string]string{})

	code, res, _ = TestAPICall(http.MethodPost, "/admin/site", b, SetupSiteRoute, "", key)
	assert.Equal(t, http.StatusBadRequest, code)

	// now update it
	input := map[string]interface{}{
		"name":        "My Site",
		"description": "My site description",
		"firstName":   "Kevin",
		"lastName":    "Eaton",
		"email":       fmt.Sprintf("r-%d@treelightsoftware.com", randID),
		"username":    fmt.Sprintf("rand-%d", randID),
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
