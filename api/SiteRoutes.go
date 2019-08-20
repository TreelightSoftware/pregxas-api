package api

import (
	"net/http"

	"github.com/go-chi/render"
)

// SiteSetup is the input for the POST to setup a new site
type SiteSetup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

// Bind binds the data
func (data *SiteSetup) Bind(r *http.Request) error {
	return nil
}

// GetSiteInfoRoute gets the site info
func GetSiteInfoRoute(w http.ResponseWriter, r *http.Request) {
	// first, check for the key
	foundKey := r.Header.Get("X-API-SECRET")
	if foundKey == "" {
		SendError(w, 400, "site_secret_key_missing", "secret key for site must be sent in X-API-SECRET header", nil)
		return
	}
	if foundKey != Site.SecretKey {
		SendError(w, 403, "site_secret_key_incorrect", "secret key incorrect", nil)
		return
	}

	key := Site.SecretKey
	Site.SecretKey = ""
	Send(w, 200, Site)
	Site.SecretKey = key
	return
}

// SetupSiteRoute sets up the site route
func SetupSiteRoute(w http.ResponseWriter, r *http.Request) {
	// first, check for the key
	foundKey := r.Header.Get("X-API-SECRET")
	if foundKey == "" {
		SendError(w, 400, "site_secret_key_missing", "secret key for site must be sent in X-API-SECRET header", nil)
		return
	}
	if foundKey != Site.SecretKey {
		SendError(w, 403, "site_secret_key_incorrect", "secret key incorrect", nil)
		return
	}

	input := SiteSetup{}
	render.Bind(r, &input)
	input.Name, _ = sanitize(input.Name)
	input.Description, _ = sanitize(input.Description)
	input.FirstName, _ = sanitize(input.FirstName)
	input.LastName, _ = sanitize(input.LastName)
	input.Email, _ = sanitize(input.Email)
	input.Username, _ = sanitize(input.Username)

	if input.Name == "" ||
		input.Description == "" ||
		input.FirstName == "" ||
		input.LastName == "" ||
		input.Email == "" ||
		input.Username == "" ||
		input.Password == "" {
		SendError(w, 400, "site_setup_invalid", "name, description, firstName, lastName, email, username, and password are all required", nil)
		return
	}
	input.Password, _ = encrypt(input.Password)

	// save the site
	site := SiteStruct{
		Name:        input.Name,
		Description: input.Description,
		Status:      "active",
		SecretKey:   "",
	}
	err := UpdateSiteSettings(&site)
	if err != nil {
		SendError(w, 400, "site_setup_invalid", "could not save site settings", err)
		return
	}

	// now the user
	user := User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  input.Password,
		Username:  input.Username,
		Status:    UserStatusVerified,
	}
	err = CreateUser(&user)
	if err != nil {
		SendError(w, 400, "site_setup_invalid", "could not create thta user", err)
		return
	}
	Send(w, 200, map[string]bool{
		"active": true,
	})
	return
}
