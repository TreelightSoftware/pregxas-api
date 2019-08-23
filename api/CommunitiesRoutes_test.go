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

func TestCommunityCRUDRoutes(t *testing.T) {
	ConfigSetup()
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	user := User{}
	err := CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user)

	user2 := User{}
	err = CreateTestUser(&user2)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user2)

	input := Community{}

	b.Reset()
	enc.Encode(&input)
	code, _, _ := TestAPICall(http.MethodPost, "/communities", b, CreateCommunityRoute, user.JWT, "")
	assert.Equal(t, http.StatusBadRequest, code)

	input = Community{
		Name: "Test Community",
	}
	b.Reset()
	enc.Encode(&input)
	code, res, _ := TestAPICall(http.MethodPost, "/communities", b, CreateCommunityRoute, user.JWT, "")
	assert.Equal(t, http.StatusCreated, code)
	_, body, _ := UnmarshalTestMap(res)
	community := Community{}
	mapstructure.Decode(body, &community)
	assert.Equal(t, input.Name, community.Name)
	assert.Equal(t, "private", community.Privacy)
	defer DeleteCommunity(community.ID)

	// update it
	update := Community{
		Name:        "A Great Group",
		Description: "A great test description",
		Privacy:     "public",
	}
	b.Reset()
	enc.Encode(&update)
	code, res, _ = TestAPICall(http.MethodPatch, fmt.Sprintf("/communities/%d", community.ID), b, UpdateCommunityRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d", community.ID), b, GetCommunityByIDRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, body, _ = UnmarshalTestMap(res)
	foundCommunity := Community{}
	mapstructure.Decode(body, &foundCommunity)
	assert.Equal(t, update.Name, foundCommunity.Name)
	assert.Equal(t, update.Description, foundCommunity.Description)
	assert.Equal(t, "public", foundCommunity.Privacy)

	// get the lists
	code, res, _ = TestAPICall(http.MethodGet, "/communities", b, GetCommunitiesForUserRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ := UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities", b, GetCommunitiesForUserRoute, user2.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Zero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities/public", b, GetCommunitiesForUserRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities/public", b, GetCommunitiesForUserRoute, user2.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.NotZero(t, len(bodyA))

	// delete and get again
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/communities/%d", community.ID), b, DeleteCommunityRoute, user2.JWT, "")
	assert.Equal(t, http.StatusForbidden, code)
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/communities/%d", community.ID), b, DeleteCommunityRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)

	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d", community.ID), b, GetCommunityByIDRoute, user.JWT, "")
	assert.Equal(t, http.StatusForbidden, code)

	code, res, _ = TestAPICall(http.MethodGet, "/communities", b, GetCommunitiesForUserRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Zero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities", b, GetCommunitiesForUserRoute, user2.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Zero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities/public", b, GetCommunitiesForUserRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Zero(t, len(bodyA))
	code, res, _ = TestAPICall(http.MethodGet, "/communities/public", b, GetCommunitiesForUserRoute, user2.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	assert.Zero(t, len(bodyA))

}

func TestCommunityLinkRoutes(t *testing.T) {
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

	user2 := User{}
	err = CreateTestUser(&user2)
	assert.Nil(t, err)
	defer DeleteUserFromTest(&user2)

	input := Community{
		Name:             "Test Community",
		Privacy:          "public",
		UserSignupStatus: "approval_required",
	}
	b.Reset()
	enc.Encode(&input)
	code, res, _ := TestAPICall(http.MethodPost, "/communities", b, CreateCommunityRoute, admin.JWT, "")
	assert.Equal(t, http.StatusCreated, code)
	_, body, _ := UnmarshalTestMap(res)
	community := Community{}
	mapstructure.Decode(body, &community)
	assert.Equal(t, input.Name, community.Name)
	assert.Equal(t, "public", community.Privacy)
	assert.Equal(t, "approval_required", community.UserSignupStatus)
	defer DeleteCommunity(community.ID)

	// now user requests to join
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/communities/%d/users/%d", community.ID, user.ID), b, RequestCommunityMembershipRoute, user.JWT, "")
	assert.Equal(t, http.StatusOK, code)

	// get the links for the community
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d/users", community.ID), b, GetCommunityLinksRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ := UnmarshalTestArray(res)
	// find the user
	found := false
	for i := range bodyA {
		link := CommunityUserLink{}
		mapstructure.Decode(bodyA[i], &link)
		if link.UserID == user.ID && link.Status == "requested" {
			found = true
		}
	}
	assert.True(t, found)

	// sanity check in the DB as well
	userLink, _ := GetCommunityUserLink(community.ID, user.ID)
	require.Equal(t, "requested", userLink.Status)

	// the admin wants user 2 to join
	code, res, _ = TestAPICall(http.MethodPut, fmt.Sprintf("/communities/%d/users/%d", community.ID, user2.ID), b, RequestCommunityMembershipRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	// get the links for the community
	code, res, _ = TestAPICall(http.MethodGet, fmt.Sprintf("/communities/%d/users", community.ID), b, GetCommunityLinksRoute, admin.JWT, "")
	assert.Equal(t, http.StatusOK, code)
	_, bodyA, _ = UnmarshalTestArray(res)
	// find the user
	found = false
	for i := range bodyA {
		link := CommunityUserLink{}
		mapstructure.Decode(bodyA[i], &link)
		if link.UserID == user2.ID && link.Status == "invited" {
			found = true
		}
	}
	assert.True(t, found)
	userLink, _ = GetCommunityUserLink(community.ID, user2.ID)
	require.Equal(t, "invited", userLink.Status)

	// admin approves user
	userLink, _ = GetCommunityUserLink(community.ID, user.ID)
	send := map[string]string{
		"shortCode": userLink.ShortCode,
		"status":    "accepted",
	}

	b.Reset()
	enc.Encode(&send)
	code, res, _ = TestAPICall(http.MethodPost, fmt.Sprintf("/communities/%d/users/%d", community.ID, user.ID), b, ProcessCommunityMembershipRoute, admin.JWT, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	userLink, _ = GetCommunityUserLink(community.ID, user.ID)
	assert.Equal(t, "member", userLink.Role)
	assert.Equal(t, "accepted", userLink.Status)

	// user2 accepts the invite
	userLink, _ = GetCommunityUserLink(community.ID, user2.ID)
	send = map[string]string{
		"shortCode": userLink.ShortCode,
		"status":    "accepted",
	}

	b.Reset()
	enc.Encode(&send)
	code, res, _ = TestAPICall(http.MethodPost, fmt.Sprintf("/communities/%d/users/%d", community.ID, user2.ID), b, ProcessCommunityMembershipRoute, user2.JWT, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	userLink, _ = GetCommunityUserLink(community.ID, user2.ID)
	assert.Equal(t, "member", userLink.Role)
	assert.Equal(t, "accepted", userLink.Status)

	// admin decided that user2 shouldn't be here anymore
	code, res, _ = TestAPICall(http.MethodDelete, fmt.Sprintf("/communities/%d/users/%d", community.ID, user2.ID), b, ProcessCommunityMembershipRoute, admin.JWT, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	userLink, err = GetCommunityUserLink(community.ID, user2.ID)
	assert.NotNil(t, err)
	assert.NotEqual(t, "member", userLink.Role)
	assert.NotEqual(t, "accepted", userLink.Status)
}
