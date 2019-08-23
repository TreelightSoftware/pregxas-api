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

func TestPrayerRequestCRUDRoutes(t *testing.T) {
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

	// get it by the id
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/requests/%d", request.ID), b, GetPrayerRequestByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundRequest := PrayerRequest{}
	mapstructure.Decode(body, &foundRequest)
	assert.Equal(t, requestInput.Title, foundRequest.Title)
	assert.Equal(t, requestInput.Body, foundRequest.Body)
	assert.Equal(t, "public", foundRequest.Privacy)
	assert.Equal(t, "pending", foundRequest.Status)

	// update it then get it again
	updateInput := map[string]string{
		"status": "answered",
	}
	b.Reset()
	enc.Encode(&updateInput)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/requests/%d", request.ID), b, UpdatePrayerRequestRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/requests/%d", request.ID), b, GetPrayerRequestByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundRequest2 := PrayerRequest{}
	mapstructure.Decode(body, &foundRequest2)
	assert.Equal(t, requestInput.Title, foundRequest2.Title)
	assert.Equal(t, requestInput.Body, foundRequest2.Body)
	assert.Equal(t, "public", foundRequest2.Privacy)
	assert.Equal(t, "answered", foundRequest2.Status)

	// add it to the community
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/communities/%d/requests/%d", community.ID, request.ID), b, AddPrayerRequestToCommunityRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	// get the global list
	code, res, _ = TestAPICall(http.MethodGet, "/requests", b, GetGlobalPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ := UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found := false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.True(t, found)

	// get the community list
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d/requests", community.ID), b, GetCommunityPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.True(t, found)

	// get for user
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/users/%d/requests", admin.ID), b, GetCommunityPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.True(t, found)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/users/%d/requests", admin.ID), b, GetCommunityPrayerRequestsRoute, user.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.True(t, found)

	// create a private prayer in the community and do the gets again

	requestInput2 := PrayerRequest{
		Title:   "Test Request",
		Body:    "Please pray for this to work",
		Privacy: "private",
	}

	b.Reset()
	enc.Encode(&requestInput2)
	code, res, _ = TestAPICall(http.MethodPost, "/requests", b, CreatePrayerRequestRoute, admin.JWT, "")
	require.Equal(t, http.StatusCreated, code)
	_, body, _ = UnmarshalTestMap(res)
	request2 := PrayerRequest{}
	mapstructure.Decode(body, &request2)
	assert.Equal(t, requestInput2.Title, request2.Title)
	assert.Equal(t, requestInput2.Body, request2.Body)
	assert.Equal(t, "private", request2.Privacy)
	defer DeletePrayerRequest(request2.ID)

	// add it to the community
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/communities/%d/requests/%d", community.ID, request2.ID), b, AddPrayerRequestToCommunityRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	// get the global list
	code, res, _ = TestAPICall(http.MethodGet, "/requests", b, GetGlobalPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request2.ID {
			found = true
		}
	}
	assert.False(t, found)

	// get the community list
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d/requests", community.ID), b, GetCommunityPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request2.ID {
			found = true
		}
	}
	assert.True(t, found)

	// get for user
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/users/%d/requests", admin.ID), b, GetCommunityPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request2.ID {
			found = true
		}
	}
	assert.True(t, found)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/users/%d/requests", admin.ID), b, GetCommunityPrayerRequestsRoute, user.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request2.ID {
			found = true
		}
	}
	assert.False(t, found)

	// finally, delete it and make sure it is gone
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/communities/%d/requests/%d", community.ID, request.ID), b, RemovePrayerRequestFromCommunityRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, "/requests", b, GetGlobalPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.True(t, found)

	// get the community list
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d/requests", community.ID), b, GetCommunityPrayerRequestsRoute, admin.JWT, "")
	require.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	// loop and find
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.False(t, found)

	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/requests/%d", request.ID), b, DeletePrayerRequestRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	code, res, _ = TestAPICall(http.MethodGet, "/requests", b, GetGlobalPrayerRequestsRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	found = false
	for i := range bodyA {
		req := PrayerRequest{}
		mapstructure.Decode(bodyA[i], &req)
		if req.ID == request.ID {
			found = true
		}
	}
	assert.False(t, found)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/requests/%d", request.ID), b, GetPrayerRequestByIDRoute, admin.JWT, "")
	require.Equal(t, http.StatusForbidden, code)
}