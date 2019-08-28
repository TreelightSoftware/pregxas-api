package api

// SiteStruct is the main installation of the site and is primarily used for standalone installations
type SiteStruct struct {
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	SecretKey    string `json:"secretKey,omitempty" db:"secretKey"`
	Status       string `json:"status" db:"status"`
	LogoLocation string `json:"logoLocation" db:"logoLocation"`
	Loaded       bool   `json:"-"`
}

// Site is the global Site variable with global configuration options from the DB
var Site = SiteStruct{}

// LoadSite lodas the site from the DB
func LoadSite() error {
	// if the site has already loaded, just return nil
	if Site.Loaded {
		return nil
	}
	err := Config.DbConn.Get(&Site, "SELECT * FROM Site LIMIT 1")
	return err
}

// SetupInitialSite sets up the initial site for a first time installation
func SetupInitialSite(secretKey string) error {
	Config.DbConn.Exec("DELETE FROM Site")
	Site.Name = "Pregxas"
	Site.Description = "The prayerful community"
	Site.SecretKey = secretKey
	Site.Status = "pending_setup"
	Site.LogoLocation = ""
	_, err := Config.DbConn.NamedExec("INSERT INTO Site (name, description, secretKey, status, logoLocation) VALUES (:name, :description, :secretKey, :status, :logoLocation)", &Site)
	return err
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
