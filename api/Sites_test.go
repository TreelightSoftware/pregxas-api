package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSite(t *testing.T) {
	ConfigSetup()
	key := GenerateSiteKey()
	assert.NotEqual(t, "", key)
	err := SetupInitialSite(key)
	assert.Nil(t, err)
	defer DeleteSiteForTest()

	assert.Equal(t, "Pregxas", Site.Name)
	assert.Equal(t, key, Site.SecretKey)
	assert.Equal(t, "pending_setup", Site.Status)

	update := SiteStruct{
		Name:        "Testing",
		Description: "Testing Desc",
		Status:      "active",
		SecretKey:   "",
	}
	err = UpdateSiteSettings(&update)
	assert.Nil(t, err)

	assert.Equal(t, "Testing", Site.Name)
	assert.Equal(t, "", Site.SecretKey)
	assert.Equal(t, "active", Site.Status)
}
