package api

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagCRD(t *testing.T) {
	ConfigSetup()
	randID := rand.Int63n(99999999)
	request1ID := rand.Int63n(999999999)
	request2ID := rand.Int63n(999999999)
	tag1 := fmt.Sprintf("teSt-1-%d", randID)
	tag2 := fmt.Sprintf("TeSt-2-%d", randID)

	created1, err := AddTagToPrayerRequest(request1ID, tag1)
	assert.Nil(t, err)
	assert.NotZero(t, created1.ID)
	assert.Equal(t, strings.ToLower(tag1), created1.Tag)

	created2, err := AddTagToPrayerRequest(request1ID, tag2)
	assert.Nil(t, err)
	assert.NotZero(t, created2.ID)
	assert.Equal(t, strings.ToLower(tag2), created2.Tag)

	// add the first tag to request 2, make sure it's the same id
	created3, err := AddTagToPrayerRequest(request2ID, tag1)
	assert.Nil(t, err)
	assert.NotZero(t, created3.ID)
	assert.Equal(t, strings.ToLower(tag1), created3.Tag)

	// defer here just in case
	defer DeleteTag(created1.ID)
	defer DeleteTag(created2.ID)

	found1, err := GetTagsOnRequest(request1ID)
	assert.Nil(t, err)
	require.Equal(t, 2, len(found1))
	assert.Equal(t, strings.ToLower(tag1), found1[0].Tag)
	assert.Equal(t, strings.ToLower(tag2), found1[1].Tag)

	found2, err := GetTagsOnRequest(request2ID)
	assert.Nil(t, err)
	require.Equal(t, 1, len(found2))
	assert.Equal(t, strings.ToLower(tag1), found2[0].Tag)

	// find one on its own
	foundSingle, err := GetTagIDByTag(tag1)
	assert.Nil(t, err)
	assert.Equal(t, created1.ID, foundSingle.ID)
	assert.Equal(t, strings.ToLower(tag1), foundSingle.Tag)

	// remove from request 1
	err = RemoveTagFromRequest(request1ID, created1.ID)
	assert.Nil(t, err)

	// fetch for both requests to make sure it didn't delete it

	found1, err = GetTagsOnRequest(request1ID)
	assert.Nil(t, err)
	require.Equal(t, 1, len(found1))
	assert.Equal(t, strings.ToLower(tag2), found1[0].Tag)

	found2, err = GetTagsOnRequest(request2ID)
	assert.Nil(t, err)
	require.Equal(t, 1, len(found2))
	assert.Equal(t, strings.ToLower(tag1), found2[0].Tag)
}
