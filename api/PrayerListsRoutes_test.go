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

	// make some bad list create requests
	b.Reset()
	enc.Encode(map[string]string{})
	code, res, _ = TestAPICall(http.MethodPost, "/lists/requests", b, CreatePrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusBadRequest, code)
	b.Reset()
	enc.Encode(map[string]string{})
	code, res, _ = TestAPICall(http.MethodPost, "/lists/requests", b, CreatePrayerListRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	b.Reset()
	enc.Encode(map[string]string{
		"title": "everything should still fail here",
	})
	code, res, _ = TestAPICall(http.MethodPost, "/lists/requests", b, CreatePrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusBadRequest, code)

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

	// make some bad add calls
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, AddPrayerRequestToPrayerListRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodPut, "/lists/requests/a/a", b, AddPrayerRequestToPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	// add the request to it
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, AddPrayerRequestToPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	// do a bad get
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodGet, "/lists/requests/a", b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundList := PrayerList{}
	mapstructure.Decode(body, &foundList)
	assert.Equal(t, listInput.Title, foundList.Title)
	assert.Equal(t, PrayerListUpdateFrequencyNever, foundList.UpdateFrequency)
	require.NotZero(t, len(foundList.PrayerRequests))
	assert.Equal(t, foundList.PrayerRequests[0].ID, request.ID)

	// do a bad update
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/lists/requests/%d", list.ID), b, UpdatePrayerListRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodPatch, "/lists/requests/a", b, UpdatePrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

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

	// get for the user
	code, res, _ = TestAPICall(http.MethodGet, "/lists/requests/", b, GetPrayerListsForUserRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodGet, "/lists/requests/", b, GetPrayerListsForUserRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ := UnmarshalTestArray(res)
	foundLists := []PrayerList{}
	mapstructure.Decode(bodyA, &foundLists)
	require.NotZero(t, len(foundLists))
	assert.Equal(t, update.Title, foundLists[0].Title)
	assert.Equal(t, PrayerListUpdateFrequencyDaily, foundLists[0].UpdateFrequency)

	// make some bad remove calls
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, RemovePrayerRequestFromPrayerListRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodDelete, "/lists/requests/a/a", b, RemovePrayerRequestFromPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	// remove the request and make sure it is gone
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d/%d", list.ID, request.ID), b, RemovePrayerRequestFromPrayerListRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundList = PrayerList{}
	mapstructure.Decode(body, &foundList)
	require.Zero(t, len(foundList.PrayerRequests))

	// bad deletes
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d", list.ID), b, DeletePrayerListByIDRoute, "", "")
	require.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodDelete, "/lists/requests/a", b, DeletePrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

	// delete the list and make sure it is gone
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/lists/requests/%d", list.ID), b, DeletePrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/lists/requests/%d", list.ID), b, GetPrayerListByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)

}
