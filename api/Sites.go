package api

// SiteStruct is the main installation of the site and is primarily used for standalone installations
type SiteStruct struct {
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	SecretKey    string `json:"secretKey" db:"secretKey"`
	Status       string `json:"status" db:"status"`
	LogoLocation string `json:"logoLocation" db:"logoLocation"`
}

// Site is the global Site variable with global configuration options from the DB
var Site = SiteStruct{}

// SetupInitialSite sets up the initial site for a first time installation
func SetupInitialSite(secretKey string) error {
	Site.Name = "Pregxas"
	Site.Description = "The prayerful community"
	Site.SecretKey = secretKey
	Site.Status = "pending_setup"
	Site.LogoLocation = ""
	return UpdateSiteSettings(&Site)
}

// DeleteSiteForTest deletes a site's settings, should only be used in testing
func DeleteSiteForTest() error {
	_, err := Config.DbConn.Exec("DELETE FROM Site")
	if err != nil {
		return err
	}
	Site = SiteStruct{}
	return nil
}

// UpdateSiteSettings updates the settings for a site
func UpdateSiteSettings(input *SiteStruct) error {
	_, err := Config.DbConn.NamedExec("UPDATE Site SET name = :name, description = :description, secretKey = :secretKey, status = :status, logoLocation = :logoLocation", input)
	if err != nil {
		return err
	}
	Site = *input
	return nil
}
