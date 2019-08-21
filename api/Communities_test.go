package api

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunityCRUD(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(999999999)
	user := User{}
	err := CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUser(user.ID)

	community := Community{
		Name:      fmt.Sprintf("Test_%d", randID),
		ShortCode: fmt.Sprintf("abc_%d", rand.Int63n(99999)),
	}

	err = CreateCommunity(&community)
	assert.Nil(t, err)
	assert.NotZero(t, community.ID)
	defer DeleteCommunity(community.ID)

	found, err := GetCommunityByID(community.ID)
	assert.Nil(t, err)
	assert.Equal(t, community.Name, found.Name)
	assert.Equal(t, community.ShortCode, found.ShortCode)

	// update it
	update := Community{
		ID:               community.ID,
		Name:             community.Name + "test",
		ShortCode:        community.ShortCode + "test",
		UserSignupStatus: CommunityUserSignupStatusAccept,
	}

	err = UpdateCommunity(&update)
	assert.Nil(t, err)
	found, err = GetCommunityByID(community.ID)
	assert.Nil(t, err)
	assert.Equal(t, community.ID, found.ID)
	assert.Equal(t, update.ID, found.ID)
	assert.Equal(t, update.Name, found.Name)
	assert.Equal(t, update.ShortCode, found.ShortCode)
	assert.Equal(t, update.UserSignupStatus, found.UserSignupStatus)

	// find by the short code
	foundShort, err := GetCommunityByShortCode(update.ShortCode)
	assert.Nil(t, err)
	assert.Equal(t, community.ID, foundShort.ID)

	// delete it, make sure it is gone
	err = DeleteCommunity(community.ID)
	assert.Nil(t, err)

	missing, err := GetCommunityByID(community.ID)
	assert.NotNil(t, err)
	assert.Zero(t, missing.ID)

}

func TestCommunityUserLinks(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(999999999)
	user := User{}
	err := CreateTestUser(&user)
	assert.Nil(t, err)
	defer DeleteUser(user.ID)

	community := Community{
		Name:      fmt.Sprintf("Test_%d", randID),
		ShortCode: fmt.Sprintf("abc_%d", rand.Int63n(99999)),
		Created:   "2019-06-09T12:45:00",
	}

	err = CreateCommunity(&community)
	assert.Nil(t, err)
	assert.NotZero(t, community.ID)
	defer DeleteCommunity(community.ID)

	// join them
	err = CreateCommunityUserLink(community.ID, user.ID, "", "invited", "abc") // blank should default to member
	assert.Nil(t, err)

	// find it
	comms, err := GetCommunitiesForUser(user.ID)
	assert.Nil(t, err)
	require.NotZero(t, len(comms))
	assert.Equal(t, community.ID, comms[0].ID)
	assert.Equal(t, "member", comms[0].UserRole)
	assert.Equal(t, CommunityUserLinkStatusInvited, comms[0].UserStatus)

	// update it
	err = UpdateCommunityUserLink(community.ID, user.ID, CommunityUserLinkStatusAccepted)
	assert.Nil(t, err)
	comms, err = GetCommunitiesForUser(user.ID)
	assert.Nil(t, err)
	require.NotZero(t, len(comms))
	assert.Equal(t, community.ID, comms[0].ID)
	assert.Equal(t, "member", comms[0].UserRole)
	assert.Equal(t, CommunityUserLinkStatusAccepted, comms[0].UserStatus)

	// get the users in the community
	users, err := GetUsersInCommunity(community.ID)
	assert.Nil(t, err)
	require.NotZero(t, len(users))
	assert.Equal(t, user.ID, users[0].ID)

	// get the count helper
	count, err := GetCountOfUsersInCommunity(community.ID)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	// delete it, make sure it is gone
	err = DeleteCommunityUserLink(community.ID, user.ID)
	assert.Nil(t, err)
	comms, err = GetCommunitiesForUser(user.ID)
	assert.Nil(t, err)
	require.Zero(t, len(comms))

}
