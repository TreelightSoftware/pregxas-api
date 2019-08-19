package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/goware/cors"
)

type key string

// AppContextKeyUser is used for the context to find the user by jwt
const AppContextKeyUser key = "user"

// AppContextKeyFound is used for the context to find the user by jwt
const AppContextKeyFound key = "found"

// PregxasAPIReturn represents a standard API return object
type PregxasAPIReturn struct {
	Data interface{} `json:"data,omitempty"`
}

// PregxasAPIError represents the data key of an error for the API
type PregxasAPIError struct {
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Send standardizes the return from the API
func Send(w http.ResponseWriter, code int, payload interface{}) {
	ret := PregxasAPIReturn{}
	ret.Data = payload
	response, _ := json.Marshal(ret)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// SendError sends an error to the client
func SendError(w http.ResponseWriter, status int, systemCode string, message string, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}
	Send(w, status, PregxasAPIError{
		Code:    systemCode,
		Message: message,
		Data:    data,
	})
}

// GetMiddlewares loads up all of the middleware needed. In order, it loads:
// - tollbooth
// - RequestID
// - RealIP
// - Logger
// - Recoverer
// - Content Type
// - Timeout (120 seconds)
// - CORS
//
func GetMiddlewares() []func(http.Handler) http.Handler {
	limiter := tollbooth.NewLimiter(Config.RateLimit, nil)
	handlers := []func(http.Handler) http.Handler{}

	h := tollbooth_chi.LimitHandler(limiter)
	handlers = append(handlers, h)

	h = middleware.RequestID
	handlers = append(handlers, h)

	h = middleware.RealIP
	handlers = append(handlers, h)

	// logging on or off?
	h = middleware.Logger
	handlers = append(handlers, h)

	h = middleware.Recoverer
	handlers = append(handlers, h)

	handlers = append(handlers, render.SetContentType(render.ContentTypeJSON))

	handlers = append(handlers, middleware.Timeout(120*time.Second))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Types", "X-CSRF-TOKEN", "Access-Control-Request-Headers", "JWT", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	h = cors.Handler
	handlers = append(handlers, h)

	handlers = append(handlers, JWTMiddleware)

	return handlers
}

// JWTMiddleware reads the JWT if it is present and attempts to parse the user. It is then
// added to the context of the HTTP request
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		found := false
		user := JWTUser{}
		key := r.Header.Get("JWT")
		var err error
		if key == "" {
			// see if it's all lower
			key = r.Header.Get("jwt")
		}

		if key != "" {
			user, err = parseJwt(key)
			if err == nil && user.ID != 0 {
				found = true
			}
		} else {
			// it may be in a cookie
			cookie, err := r.Cookie("jwt")
			if err == nil && cookie != nil && cookie.Value != "" {
				user, err = parseJwt(cookie.Value)
				if err == nil && user.ID != 0 {
					found = true
				}
			}
		}

		ctx := context.WithValue(r.Context(), AppContextKeyFound, found)
		ctx = context.WithValue(ctx, AppContextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CheckForUser is a helper to check the context for a user. It checks if the users is an admin. If not, it checks if the user
// has the needed permission
func CheckForUser(r *http.Request) (user JWTUser, err error) {
	user = JWTUser{}
	found := r.Context().Value(AppContextKeyFound).(bool)
	user, userOK := r.Context().Value(AppContextKeyUser).(JWTUser)
	if !found || !userOK {
		err = errors.New("could not parse JWT")
		return
	}
	return
}

// ProcessQuery parses the query string tokens for the following fields and then returns them:
// start - The start of a date filter
// end - The end of a date filter
// limit - The number of entries to return
// offset - Any offset in the pagination
// sortField - Any field that should be sorted
// sortDir - The direction of the sort
// filterKey - The filter field
// filterValue - The value to filter on
func ProcessQuery(r *http.Request) (start, end string, count, offset int, sortField, sortDir, filterKey, filterValue string) {
	startQ := r.URL.Query().Get("start")
	endQ := r.URL.Query().Get("end")
	countQ := r.URL.Query().Get("count")
	offsetQ := r.URL.Query().Get("offset")
	sortDirQ := r.URL.Query().Get("sortDir")
	sortField = r.URL.Query().Get("sortField")
	filterKey = r.URL.Query().Get("filterKey")
	filterValue = r.URL.Query().Get("filterValue")

	start, err := ParseISOTimeToDBTime(startQ)
	if err != nil {
		start = "2017-01-01 00:00:00"
	}
	end, err = ParseISOTimeToDBTime(endQ)
	if err != nil {
		end = "2020-01-01 00:00:00"
	}

	//try to convert the limit and offset
	count, err = strconv.Atoi(countQ)
	if err != nil {
		count = 500
	}

	offset, err = strconv.Atoi(offsetQ)
	if err != nil {
		offset = 0
	}

	sortDir = strings.ToUpper(sortDirQ)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}
	return
}
