package api

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrayerListCRUD(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999999)
	request := PrayerRequest{
		Title:     fmt.Sprintf("Test Prayer %d", randID),
		Body:      "Test Prayer Request Body",
		CreatedBy: randID,
		Privacy:   "public",
	}
	err := CreatePrayerRequest(&request)
	require.Nil(t, err)
	defer DeletePrayerRequest(request.ID)

	list := PrayerList{
		UserID:          randID,
		Title:           "Test List",
		UpdateFrequency: PrayerListUpdateFrequencyDaily,
	}

	err = CreatePrayerList(&list)
	require.Nil(t, err)
	require.NotZero(t, list.ID)
	defer DeletePrayerList(list.ID)

	found, err := GetPrayerList(list.ID)
	assert.Nil(t, err)
	assert.Equal(t, list.Title, found.Title)
	assert.Zero(t, len(list.PrayerRequests))
	assert.Equal(t, PrayerListUpdateFrequencyDaily, found.UpdateFrequency)

	list.UpdateFrequency = PrayerListUpdateFrequencyWeekly
	list.Title = "Updated"
	err = UpdatePrayerList(&list)
	assert.Nil(t, err)

	found, err = GetPrayerList(list.ID)
	assert.Nil(t, err)
	assert.Equal(t, list.Title, found.Title)
	assert.Zero(t, len(list.PrayerRequests))
	assert.Equal(t, PrayerListUpdateFrequencyWeekly, found.UpdateFrequency)

	// now add a request to the list
	err = AddRequestToPrayerList(request.ID, list.ID)
	assert.Nil(t, err)

	requests, err := GetPrayerRequestsOnPrayerList(list.ID)
	assert.Nil(t, err)
	assert.NotZero(t, len(requests))
	assert.Equal(t, request.ID, requests[0].ID)

	// get for the users
	lists, err := GetPrayerListsForUser(randID, "title")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(lists))
	lists, err = GetPrayerListsForUser(randID, "createdOn")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(lists))

	// remove and verify
	err = RemoveRequestFromPrayerList(request.ID, list.ID)
	assert.Nil(t, err)

	requests, err = GetPrayerRequestsOnPrayerList(list.ID)
	assert.Nil(t, err)
	assert.Zero(t, len(requests))

}
