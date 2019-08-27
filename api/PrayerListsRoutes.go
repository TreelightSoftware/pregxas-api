package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Bind binds data
func (data *PrayerList) Bind(r *http.Request) error {
	return nil
}

// CreatePrayerListRoute creates a new prayer list
func CreatePrayerListRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := PrayerList{}
	render.Bind(r, &input)
	input.UserID = jwtUser.ID
	input.Title, _ = sanitize(input.Title)
	input.UpdateFrequency = strings.ToLower(input.UpdateFrequency)

	if input.Title == "" {
		SendError(w, http.StatusBadRequest, "prayer_request_list_create_bad_data", "title is required", input)
		return
	}
	if input.UpdateFrequency != PrayerListUpdateFrequencyDaily &&
		input.UpdateFrequency != PrayerListUpdateFrequencyNever &&
		input.UpdateFrequency != PrayerListUpdateFrequencyWeekly {
		SendError(w, http.StatusBadRequest, "prayer_request_list_create_bad_data", "list updateFrequency must be either 'daily', 'weekly' or 'never'", input)
		return
	}

	err = CreatePrayerList(&input)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_list_create_bad_data", "could not create that list", err)
		return
	}

	Send(w, http.StatusCreated, input)
	return
}

// GetPrayerListsForUserRoute gets all of the lists for a user
func GetPrayerListsForUserRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	_, _, _, _, sortField, _, _, _ := ProcessQuery(r)

	lists, err := GetPrayerListsForUser(jwtUser.ID, sortField)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	Send(w, http.StatusOK, lists)
	return
}

// GetPrayerListByIDRoute gets a single list for a user
func GetPrayerListByIDRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	listID, listIDErr := strconv.ParseInt(chi.URLParam(r, "listID"), 10, 64)
	if listIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	list, err := GetPrayerList(listID)
	if err != nil || list.UserID != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	Send(w, http.StatusOK, list)
	return
}

// UpdatePrayerListRoute updates the title or updateFrequency of a list
func UpdatePrayerListRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	listID, listIDErr := strconv.ParseInt(chi.URLParam(r, "listID"), 10, 64)
	if listIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	list, err := GetPrayerList(listID)
	if err != nil || list.UserID != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := PrayerList{}
	render.Bind(r, &input)
	input.Title, _ = sanitize(input.Title)

	if input.Title != "" {
		list.Title = input.Title
	}

	if input.UpdateFrequency == PrayerListUpdateFrequencyDaily ||
		input.UpdateFrequency == PrayerListUpdateFrequencyNever ||
		input.UpdateFrequency == PrayerListUpdateFrequencyWeekly {
		list.UpdateFrequency = input.UpdateFrequency
	}

	err = UpdatePrayerList(&list)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_list_update_error", "could not update that list", err)
		return
	}

	Send(w, http.StatusOK, list)
	return
}

// DeletePrayerListByIDRoute deletes a single list
func DeletePrayerListByIDRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	listID, listIDErr := strconv.ParseInt(chi.URLParam(r, "listID"), 10, 64)
	if listIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	list, err := GetPrayerList(listID)
	if err != nil || list.UserID != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	err = DeletePrayerList(listID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_list_delete_error", "could not delete that prayer request list", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
	return
}

// AddPrayerRequestToPrayerListRoute adds a request to a list
func AddPrayerRequestToPrayerListRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	listID, listIDErr := strconv.ParseInt(chi.URLParam(r, "listID"), 10, 64)
	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if listIDErr != nil || requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	list, err := GetPrayerList(listID)
	if err != nil || list.UserID != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	request, err := GetPrayerRequest(requestID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// ok, so there's some interesting logic here
	// first, if the request belongs to the user or is public, it's good
	// if not, if the request is in a group that the user belongs to, then it too is good
	canAdd := false
	if request.Privacy == "public" || request.CreatedBy == jwtUser.ID || IsUserAndRequestInSameGroup(jwtUser.ID, requestID) {
		canAdd = true
	}

	if !canAdd {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	err = AddRequestToPrayerList(requestID, listID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_list_add_error", "could not add that prayer request to that list", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"added": true,
	})
	return
}

// RemovePrayerRequestFromPrayerListRoute removes a request from a list
func RemovePrayerRequestFromPrayerListRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	listID, listIDErr := strconv.ParseInt(chi.URLParam(r, "listID"), 10, 64)
	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if listIDErr != nil || requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	list, err := GetPrayerList(listID)
	if err != nil || list.UserID != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// the logic is fairly simple; if they own the list they can remove the request, regardless if the request was changed to private, etc
	err = RemoveRequestFromPrayerList(requestID, listID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_list_remove_error", "could not remove that prayer request from that list", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"removed": true,
	})
	return
}
