package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrayerRequestListCRUDRoutes(t *testing.T) {
	ConfigSetup()
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	admin := User{}
	err := CreateTestUser(&admin)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&admin)

	user := User{}
	err = CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user)

	community := Community{
		Name:             "Test Community",
		Privacy:          "public",
		UserSignupStatus: "approval_required",
	}
	err = CreateCommunity(&community)
	require.Nil(t, err)
	defer DeleteCommunity(community.ID)
	err = CreateCommunityUserLink(community.ID, admin.ID, "admin", "accepted", "")
	require.Nil(t, err)

	// create a request
	requestInput := PrayerRequest{
		Title:   "Test Request",
		Body:    "Please pray for this to work",
		Privacy: "public",
	}

	b.Reset()
	enc.Encode(&requestInput)
	code, res, _ := TestAPICall(http.MethodPost, "/requests", b, CreatePrayerRequestRoute, admin.JWT, "")
	require.Equal(t, http.StatusCreated, code)
	_, body, _ := UnmarshalTestMap(res)
	request := PrayerRequest{}
	mapstructure.Decode(body, &request)
	assert.Equal(t, requestInput.Title, request.Title)
	assert.Equal(t, requestInput.Body, request.Body)
	assert.Equal(t, "public", request.Privacy)
	defer DeletePrayerRequest(request.ID)

	// create a list for the admin user
	listInput := PrayerList{
		Title:           "My List",
		UpdateFrequency: PrayerListUpdateFrequencyNever,
	}
	b.Reset()
	enc.Encode(&listInput)
	code, res, _ = TestAPICall(http.MethodPost, "/lists/requests", b, CreatePrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusCreated, code)
	_, body, _ = UnmarshalTestMap(res)
	list := PrayerList{}
	mapstructure.Decode(body, &list)
	assert.Equal(t, listInput.Title, list.Title)
	defer DeletePrayerList(list.ID)

	// add the request to it
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, AddPrayerRequestToPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundList := PrayerList{}
	mapstructure.Decode(body, &foundList)
	assert.Equal(t, listInput.Title, foundList.Title)
	assert.Equal(t, PrayerListUpdateFrequencyNever, foundList.UpdateFrequency)
	require.NotZero(t, len(foundList.PrayerRequests))
	assert.Equal(t, foundList.PrayerRequests[0].ID, request.ID)

	// update it
	update := PrayerList{
		Title:           "Updated",
		UpdateFrequency: PrayerListUpdateFrequencyDaily,
	}
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/lists/requests/%d", list.ID), b, UpdatePrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundList = PrayerList{}
	mapstructure.Decode(body, &foundList)
	assert.Equal(t, update.Title, foundList.Title)
	assert.Equal(t, PrayerListUpdateFrequencyDaily, foundList.UpdateFrequency)

	// remove the request and make sure it is gone
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, RemovePrayerRequestFromPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundList = PrayerList{}
	mapstructure.Decode(body, &foundList)
	require.Zero(t, len(foundList.PrayerRequests))

	// delete the list and make sure it is gone
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d", list.ID), b, DeletePrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

}
