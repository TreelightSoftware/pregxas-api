package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type refreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

// Bind binds the data for the HTTP
func (data *refreshTokenInput) Bind(r *http.Request) error {
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

	// we are changing to take the jwt and use it as an access token
	// and send it to the user in an http-only cookie

	found, err := LoginUser(input.Email, input.Password)
	if err != nil {
		SendError(w, http.StatusUnauthorized, "user_invalid_data", "could not log the user in", nil)
		return
	}
	if found.Status != UserStatusVerified {
		SendError(w, http.StatusForbidden, "user_login_not_verified", "user not verified", found)
		return
	}
	// generate a JWT and send it in a header
	jwt, err := createJwt(found)
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    jwt,
		Expires:  time.Now().Add(time.Second * AccessTokenExpiresSeconds),
		MaxAge:   AccessTokenExpiresSeconds,
		Domain:   Config.RootAPIDomain,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsDev() || IsProd(),
	}
	http.SetCookie(w, accessCookie)

	// TODO: generate a refresh token
	refreshToken, err := GenerateToken(found.ID, TokenRefresh)
	if err != nil {
		// something went REALLY bad, but we will allow things to continue
		// they just won't be able to refresh their token
		Log("warning", "refresh token could not be created", "refresh_token_error", map[string]string{
			"error": err.Error(),
		})
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   0,
		Domain:   Config.RootAPIDomain,
		Path:     "/users/refresh",
		HttpOnly: true,
		Secure:   IsDev() || IsProd(),
	}
	http.SetCookie(w, refreshCookie)

	// we still return the JWT in the body of the login so that non-web integrations have access to it
	// remember to not store it in local storage!
	found.JWT = jwt
	found.AccessToken = jwt
	found.RefreshToken = refreshToken
	found.ExpiresIn = AccessTokenExpiresSeconds
	found.ExpiresAt = time.Now().Add(time.Second * AccessTokenExpiresSeconds).Format("2006-01-02T15:04:05Z")

	found.Password = ""
	Send(w, http.StatusOK, found)
	return
}

// RefreshAccessTokenRoute takes in the refresh_token from the cookie and attempts to
// refresh the access token. The refresh token can be in either a secured cookie (web) or the body (anything else).
// The return will be the same as if the user logged in
func RefreshAccessTokenRoute(w http.ResponseWriter, r *http.Request) {

	// first check the cookies
	inputToken := ""
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie != nil {
		inputToken = cookie.Value
	}
	if inputToken == "" {
		// check the POST
		input := refreshTokenInput{}
		render.Bind(r, &input)
		inputToken = input.RefreshToken
	}

	if inputToken == "" {
		// it's still missing, so error out
		SendError(w, http.StatusBadRequest, "refresh_token_missing", "missing refresh_token", nil)
		return
	}

	// ok, so now we split; a refresh token will match a specific pattern
	parts := strings.Split(inputToken, "_")
	if len(parts) == 0 {
		SendError(w, http.StatusBadRequest, "refresh_token_invalid", "the passed in refresh token is invalid", nil)
		return
	}
	// the first part should be the ruserId
	userIDString := parts[0][1:]
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "refresh_token_invalid", "the passed in refresh token is invalid", nil)
		return
	}

	// now find the token
	token, err := GetTokenForTest(userID, TokenRefresh)
	if err != nil {
		SendError(w, http.StatusBadRequest, "refresh_token_invalid", "the passed in refresh token is invalid", nil)
		return
	}

	if token != inputToken {
		SendError(w, http.StatusBadRequest, "refresh_token_invalid", "the passed in refresh token is invalid", nil)
		return
	}

	// they match, so let's relog them in with a new access token; refresh token can stay
	user, err := GetUserByID(userID)
	if err != nil {
		// well this is awkward...
		SendError(w, http.StatusForbidden, "refresh_token_user", "the user is not valid", map[string]string{
			"userID": userIDString,
			"error":  err.Error(),
		})
		return
	}
	jwt, err := createJwt(user)
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    jwt,
		Expires:  time.Now().Add(time.Second * AccessTokenExpiresSeconds),
		MaxAge:   AccessTokenExpiresSeconds,
		Domain:   Config.RootAPIDomain,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsDev() || IsProd(),
	}
	http.SetCookie(w, accessCookie)

	Send(w, http.StatusOK, map[string]interface{}{
		"access_token": jwt,
		"expires_in":   AccessTokenExpiresSeconds,
		"expires_at":   time.Now().Add(time.Second & AccessTokenExpiresSeconds).Format("2006-01-02T15:04:05Z"),
	})
	return
}

// LogoutUserRoute nukes the cookies but current does NOT delete the refresh token
func LogoutUserRoute(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now(),
		MaxAge:   -1,
		Domain:   Config.RootAPIDomain,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsDev() || IsProd(),
	}
	http.SetCookie(w, cookie)
	cookie.Name = "refresh_token"
	http.SetCookie(w, cookie)

	Send(w, http.StatusOK, map[string]bool{
		"loggedOut": true,
	})
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
