package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

type verifyUserInput struct {
	Email     string `json:"email"`
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
}

// Bind binds the data for the HTTP
func (data *verifyUserInput) Bind(r *http.Request) error {
	return nil
}

type resetPasswordInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Bind binds the data for the HTTP
func (data *resetPasswordInput) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *User) Bind(r *http.Request) error {
	return nil
}

// SignupUserRoute signs up a new user
func SignupUserRoute(w http.ResponseWriter, r *http.Request) {
	input := User{}
	render.Bind(r, &input)

	// sanitize; errors can be ignored due to the blank checks
	input.Username, _ = sanitize(strings.ToLower(input.Username))
	input.Email, _ = sanitize(strings.ToLower(input.Email))
	input.FirstName, _ = sanitize(input.FirstName)
	input.LastName, _ = sanitize(input.LastName)

	if input.Email == "" || input.FirstName == "" || input.Password == "" || input.Username == "" {
		SendError(w, http.StatusBadRequest, "invalid_data", "email, username, firstName, and password are all required", nil)
		return
	}
	// see if the user exists
	found, err := GetUserByEmail(input.Email)
	if err == nil && found.ID != 0 {
		SendError(w, http.StatusBadRequest, "user_signup_error", "there was an error when creating that user; it is possible that email address already exists", nil)
		return
	}

	found, err = GetUserByUsername(input.Username)
	if err == nil && found.ID != 0 {
		SendError(w, http.StatusBadRequest, "user_signup_error", "there was an error when creating that user; it is possible that username already exists", nil)
		return
	}

	input.Status = UserStatusPending

	err = CreateUser(&input)
	if err != nil {
		SendError(w, http.StatusBadRequest, "user_signup_error", "there was an error when creating that user; it is possible that email address already exists", nil)
		return
	}

	// generate a token and send an email
	token, _ := GenerateToken(input.ID, TokenEmailVerify)

	// TODO: break email bodies out into templates
	tokenURL := fmt.Sprintf("%sverify", Config.WebURL)
	tokenWithParams := fmt.Sprintf("%s?email=%s&token=%s", tokenURL, input.Email, token)
	emailContent := fmt.Sprintf(`<p>You, or someone pretending to be you,signed up for an account for a platform provided by <a href="%s">Pregxas</a>. 
	If this was not you, you can safely ignore this email.</p>
	<p>Otherwise, please verify your email with us by clicking <a href="%s">here</a>.</p>
	<p>If that link does not work for you, you can visit <a href="%s">%s</a> and enter the following information:<p>
	<p>Email: %s<br />Token: %s</p>
	<p>Thanks!</p>
	`, Config.WebURL, tokenWithParams, tokenURL, tokenURL, input.Email, token)

	emailBody := GenerateEmail(0, emailContent)
	SendEmail(input.Email, "Verify Your Account", emailBody)

	Send(w, 201, map[string]interface{}{
		"status": "pending",
		"id":     input.ID,
	})
}

// GetMyProfileRoute gets the current user's profile
func GetMyProfileRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	found, err := GetUserByID(jwtUser.ID)
	if err != nil || found == nil {
		// this is very weird, since the JWT is there but the user isn't?
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	found.clean()
	Send(w, http.StatusOK, found)
	return
}

// UpdateMyProfileRoute updates a student's profile
func UpdateMyProfileRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	found, err := GetUserByID(jwtUser.ID)
	if err != nil || found == nil {
		// this is very weird, since the JWT is there but the user isn't?
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := User{
		FirstName: "-1",
		LastName:  "-1",
		Email:     "-1",
		Password:  "-1",
		Username:  "-1",
	}
	render.Bind(r, &input)

	// sanitize; errors can be ignored due to the blank checks
	input.Email, _ = sanitize(strings.ToLower(input.Email))
	input.Username, _ = sanitize(strings.ToLower(input.Username))
	input.FirstName, _ = sanitize(input.FirstName)
	input.LastName, _ = sanitize(input.LastName)

	// let's find out what changed
	if input.FirstName != "-1" && input.FirstName != "" {
		found.FirstName = input.FirstName
	}
	if input.LastName != "-1" && input.LastName != "" {
		found.LastName = input.LastName
	}
	if input.Email != "-1" && input.Email != "" && input.Email != jwtUser.Email {
		foundWithEmail, _ := GetUserByEmail(input.Email)
		if foundWithEmail.ID != 0 && foundWithEmail.ID != found.ID {
			SendError(w, http.StatusBadRequest, "user_save_error", "could not update that user's information", nil)
			return
		}
		found.Email = input.Email
	}
	if input.Username != "-1" && input.Username != "" && input.Username != jwtUser.Username {
		foundWithUsername, _ := GetUserByUsername(input.Username)
		if foundWithUsername.ID != 0 && foundWithUsername.ID != found.ID {
			SendError(w, http.StatusBadRequest, "user_save_error", "could not update that user's information", nil)
			return
		}
		found.Username = input.Username
	}
	if input.Password != "-1" && input.Password != "" {
		found.Password = input.Password // the process for DB handles the encryption
	}

	err = UpdateUser(found)
	if err != nil {
		SendError(w, http.StatusBadRequest, "user_save_error", "could not update that user's information", nil)
		return
	}
	found.clean()
	Send(w, http.StatusOK, found)
	return
}

// LoginUserRoute logs a user in to the platform
func LoginUserRoute(w http.ResponseWriter, r *http.Request) {
	input := User{}
	render.Bind(r, &input)

	if input.Email == "" || input.Password == "" {
		SendError(w, http.StatusBadRequest, "user_invalid_data", "email and password are required", nil)
		return
	}
	input.Email = strings.ToLower(input.Email)

	found, err := LoginUser(input.Email, input.Password)
	if err != nil {
		SendError(w, http.StatusUnauthorized, "user_invalid_data", "could not log the user in", nil)
		return
	}
	if found.Status != UserStatusVerified {
		SendError(w, http.StatusForbidden, "user_login_not_verified", "user not verified", found)
		return
	}
	found.Password = ""
	Send(w, http.StatusOK, found)
	return
}

// VerifyEmailAndTokenRoute verifies the user account
func VerifyEmailAndTokenRoute(w http.ResponseWriter, r *http.Request) {
	input := verifyUserInput{}
	render.Bind(r, &input)

	if input.Email == "" || input.Token == "" {
		SendError(w, http.StatusBadRequest, "user_verify_blank_data", "username and token are required", input)
		return
	}
	input.Email = strings.ToLower(input.Email)

	userID, valid, err := VerifyToken(input.Token, TokenEmailVerify)
	if err != nil || !valid {
		SendError(w, http.StatusForbidden, "user_verify_failed", "could not verify that token", input)
		return
	}

	foundUser, err := GetUserByID(userID)
	if err != nil || foundUser.Email != input.Email {
		SendError(w, http.StatusForbidden, "user_verify_failed", "could not verify that token", input)
		return
	}

	foundUser.Status = UserStatusVerified
	err = UpdateUser(foundUser)
	if err != nil {
		SendError(w, http.StatusBadRequest, "user_verify_bad_update", "could not update that user", input)
		return
	}

	Send(w, http.StatusOK, map[string]interface{}{
		"verified": true,
		"id":       foundUser.ID,
	})
	return
}

// ResetPasswordStartRoute starts the password reset process
func ResetPasswordStartRoute(w http.ResponseWriter, r *http.Request) {
	input := resetPasswordInput{}
	render.Bind(r, &input)

	if input.Email == "" {
		SendError(w, http.StatusBadRequest, "reset_password_no_email", "we need an email", map[string]string{
			"email": input.Email,
		})
		return
	}

	user, err := GetUserByEmail(input.Email)
	if err != nil {
		SendError(w, http.StatusForbidden, "reset_password_forbidden", "could not reset that account", nil)
		return
	}

	token, err := GenerateToken(user.ID, TokenPasswordReset)
	if err != nil {
		SendError(w, 500, "token_error", "could not generate a reset token", map[string]string{
			"error": err.Error(),
		})
		return
	}
	tokenURL := fmt.Sprintf("%susers/login/reset/verify", Config.WebURL)
	tokenWithParams := fmt.Sprintf("%s?email=%s&token=%s", tokenURL, user.Email, token)

	emailContent := fmt.Sprintf(`<p>You, or someone pretending to be you, is resetting their password for <a href="%s">Pregxas</a>. If this was not you, you can safely ignore this email.</p>
	<p>Otherwise, please verify your email and reset your password with us by clicking <a href="%s">here</a>.</p>
	<p>If that link does not work for you, you can visit <a href="%s">%s</a> and enter the following information:<p>
	<p>Email: %s<br />Token: %s</p>
	<p>Thanks!</p>
	`, Config.WebURL, tokenWithParams, tokenURL, tokenURL, user.Email, token)

	emailBody := GenerateEmail(0, emailContent)
	SendEmail(user.Email, "Reset Your Password", emailBody)
	Send(w, http.StatusOK, map[string]bool{
		"resetStarted": true,
	})
	return
}

// ResetPasswordVerifyRoute verifies the password reset token
func ResetPasswordVerifyRoute(w http.ResponseWriter, r *http.Request) {
	input := resetPasswordInput{}
	render.Bind(r, &input)

	if input.Email == "" {
		SendError(w, http.StatusBadRequest, "reset_password_no_email", "we need your email", map[string]string{
			"username": input.Email,
		})
		return
	}

	if input.Password == "" || input.Token == "" {
		SendError(w, http.StatusBadRequest, "reset_password_no_verification", "we need the token and the password", map[string]string{})
		return
	}

	// find it
	userID, valid, err := VerifyToken(input.Token, TokenPasswordReset)
	if !valid || err != nil || userID == 0 {
		SendError(w, http.StatusForbidden, "reset_password_permission_denied", "permission denied", input)
		return
	}

	found, err := GetUserByID(userID)
	if found.Email != input.Email {
		SendError(w, http.StatusForbidden, "reset_password_permission_denied", "permission denied: emails don't match", input)
		return
	}

	found.Password = input.Password
	found.Status = UserStatusVerified

	UpdateUser(found)

	Send(w, http.StatusOK, map[string]bool{
		"passwordReset": true,
	})
	return

}
