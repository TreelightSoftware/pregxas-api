package api

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrayerCRUD(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999999)
	user1ID := rand.Int63n(99999999)
	user2ID := rand.Int63n(99999999)
	user3ID := rand.Int63n(99999999)

	community := Community{
		Name:      fmt.Sprintf("Test_%d", randID),
		ShortCode: fmt.Sprintf("abc_%d", rand.Int63n(99999)),
		Created:   "2019-06-09T12:45:00",
	}

	err := CreateCommunity(&community)
	assert.Nil(t, err)
	assert.NotZero(t, community.ID)
	defer DeleteCommunity(community.ID)

	request := PrayerRequest{
		Title:     fmt.Sprintf("Test Prayer %d", randID),
		Body:      "Test Prayer Request Body",
		CreatedBy: user1ID,
		Privacy:   "public",
	}
	err = CreatePrayerRequest(&request)
	require.Nil(t, err)
	defer DeletePrayerRequest(request.ID)

	request2 := PrayerRequest{
		Title:     fmt.Sprintf("Test Prayer %d", randID),
		Body:      "Test Prayer Request Body",
		CreatedBy: user1ID,
		Privacy:   "public",
		Created:   time.Now().Format("2006-01-02 15:04:05"),
	}
	err = CreatePrayerRequest(&request2)
	require.Nil(t, err)
	defer DeletePrayerRequest(request2.ID)

	found, err := GetPrayerRequest(request.ID)
	assert.Nil(t, err)
	assert.Equal(t, request.Title, found.Title)
	assert.Equal(t, "public", found.Privacy)
	assert.Equal(t, PrayerRequestStatusPending, found.Status)

	// add it to a community of 0 which is the main feed
	err = AddPrayerRequestToCommunity(request.ID, 0)
	assert.Nil(t, err)

	// also add it to the created community
	err = AddPrayerRequestToCommunity(request.ID, community.ID)
	assert.Nil(t, err)

	// offer a prayer from each
	err = AddPrayerMade(user1ID, request.ID)
	assert.Nil(t, err)
	err = AddPrayerMade(user2ID, request.ID)
	assert.Nil(t, err)

	count := GetCountOfPrayersMadeForRequest(request.ID, "", "")
	assert.Equal(t, int64(2), count)
	count = GetCountOfPrayersMadeForRequest(request.ID, "2000-01-01 00:00:00", "2001-01-01 00:00:00")
	assert.Equal(t, int64(0), count)

	canSubmit, minutes := CanUserMakeNewPrayer(user1ID, request.ID)
	assert.False(t, canSubmit)
	assert.NotZero(t, minutes)
	canSubmit, minutes = CanUserMakeNewPrayer(user2ID, request.ID)
	assert.False(t, canSubmit)
	assert.NotZero(t, minutes)
	canSubmit, minutes = CanUserMakeNewPrayer(user3ID, request.ID)
	assert.True(t, canSubmit)
	assert.Zero(t, minutes)

	// insert a prayer made by hand for user3ID to test the length of time check
	_, err = Config.DbConn.Exec("INSERT INTO Prayers (prayerRequestId, userId, whenPrayed) VALUES (?,?,'2001-01-02 00:00:00')", request.ID, user3ID)
	assert.Nil(t, err)
	// make sure they are still eligible
	canSubmit, minutes = CanUserMakeNewPrayer(user3ID, request.ID)
	assert.True(t, canSubmit)
	assert.Zero(t, minutes)

	// prayer was answered, update the status
	found.Status = PrayerRequestStatusAnswered
	err = UpdatePrayerRequest(found)
	assert.Nil(t, err)

	found, err = GetPrayerRequest(request.ID)
	assert.Nil(t, err)
	assert.Equal(t, PrayerRequestStatusAnswered, found.Status)

	// get the request for the user
	requests, err := GetUserPrayerRequests(user1ID, "", "", "", 1000, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(requests))
	// the second entry should have 3 prayers
	assert.Equal(t, 3, requests[1].PrayerCount)

	requests, err = GetUserPrayerRequests(user1ID, PrayerRequestStatusPending, "", "", 1000, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(requests))

	requests, err = GetUserPrayerRequests(user1ID, "", "2001-01-02 00:00:00", "2002-01-02 00:00:00", 1000, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(requests))

	// remove a prayer made and then get again to make sure it is gone
	prayers, err := GetPrayersMadeByUserForRequest(user1ID, request.ID, 10000, 0)
	assert.Nil(t, err)
	err = RemovePrayerMade(user1ID, request.ID, prayers[0].WhenPrayed)
	assert.Nil(t, err)

	found, err = GetPrayerRequest(request.ID)
	assert.Nil(t, err)
	assert.Equal(t, PrayerRequestStatusAnswered, found.Status)
	assert.Equal(t, 2, found.PrayerCount)
	canSubmit, minutes = CanUserMakeNewPrayer(user1ID, request.ID)
	assert.True(t, canSubmit)
	assert.Zero(t, minutes)

	// get the count of requests for each community
	countInCommunity, err := GetCountOfRequestsInCommunity(0, "", "")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), countInCommunity)
	countInCommunity, err = GetCountOfRequestsInCommunity(community.ID, "", "")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), countInCommunity)

	err = RemovePrayerRequestFromCommunity(request.ID, community.ID)
	assert.Nil(t, err)

}
