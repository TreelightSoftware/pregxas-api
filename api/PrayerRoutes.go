package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Bind binds data
func (data *Prayer) Bind(r *http.Request) error {
	return nil
}

// Bind binds data
func (data *PrayerRequest) Bind(r *http.Request) error {
	return nil
}

// CreatePrayerRequestRoute creates a new prayer request that can then be joined to communities
func CreatePrayerRequestRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := PrayerRequest{}
	render.Bind(r, &input)
	input.Title, _ = sanitize(input.Title)
	input.Body, _ = sanitize(input.Body)
	input.PrayerCount = 0
	input.CreatedBy = jwtUser.ID
	input.Status = PrayerRequestStatusPending
	if input.Privacy == "" {
		input.Privacy = "private"
	}

	if input.Title == "" || input.Body == "" {
		SendError(w, http.StatusBadRequest, "prayer_request_bad_data", "title and body are required", input)
		return
	}

	err = CreatePrayerRequest(&input)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_bad_data", "prayer request could not be created", err)
		return
	}

	// now process the tags
	for i := range input.Tags {
		AddTagToPrayerRequest(input.ID, input.Tags[i])
	}

	Send(w, http.StatusCreated, input)
	return
}

// UpdatePrayerRequestRoute updates the status or privacy of a prayer request. Note that once created, we do not allow for changing of titles, bodies, etc
func UpdatePrayerRequestRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// get the request to ensure permissions
	request, err := GetPrayerRequest(requestID)
	if err != nil || request.CreatedBy != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := PrayerRequest{}
	render.Bind(r, &input)

	if input.Privacy != "" {
		request.Privacy = input.Privacy
	}

	if input.Status != "" {
		request.Status = input.Status
	}
	err = UpdatePrayerRequest(request)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_bad_data", "prayer request could not be updated", err)
		return
	}

	Send(w, http.StatusOK, request)
	return
}

// DeletePrayerRequestRoute deletes a prayer request and the community links
func DeletePrayerRequestRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// get the request to ensure permissions
	request, err := GetPrayerRequest(requestID)
	if err != nil || request.CreatedBy != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	err = DeletePrayerRequest(requestID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_bad_data", "prayer request could not be deleted", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
	return
}

// GetPrayerRequestByIDRoute gets a prayer request by its id
func GetPrayerRequestByIDRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if requestIDErr != nil {
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

	if request.Privacy == "public" || request.CreatedBy == jwtUser.ID {
		Send(w, http.StatusOK, request)
		return
	}

	// check the groups
	if IsUserAndRequestInSameGroup(jwtUser.ID, requestID) {
		Send(w, http.StatusOK, request)
		return
	}

	SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
	return
}

// GetGlobalPrayerRequestsRoute gets the global request list
func GetGlobalPrayerRequestsRoute(w http.ResponseWriter, r *http.Request) {
	_, _, count, offset, _, _, _, _ := ProcessQuery(r)
	requests := GetGlobalPrayerRequests(count, offset)

	Send(w, http.StatusOK, requests)
	return
}

// GetUserPrayerRequestsRoute gets the requests for a user
func GetUserPrayerRequestsRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if userIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	status := r.URL.Query().Get("status")
	start, end, count, offset, _, _, _, _ := ProcessQuery(r)
	requests, _ := GetUserPrayerRequests(userID, status, start, end, count, offset)

	// if the userID and the jwt user are the same, show all
	if userID == jwtUser.ID {
		Send(w, http.StatusOK, requests)
		return
	}

	// since it is a different user, we only want to return requests that are public
	// in the future, we can modify the SQL to get the communities each prayer is in and filter on that
	processed := []PrayerRequest{}
	for i := range requests {
		if requests[i].Privacy == "public" {
			processed = append(processed, requests[i])
		}
	}

	Send(w, http.StatusOK, processed)
	return
}

// GetCommunityPrayerRequestsRoute gets the requests for a user
func GetCommunityPrayerRequestsRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	if communityIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// make sure the user is in the community
	link, err := GetCommunityUserLink(communityID, jwtUser.ID)
	if err != nil || link.Status != "accepted" || (link.Role != "admin" && link.Role != "member") {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	status := r.URL.Query().Get("status")
	_, _, count, offset, _, _, _, _ := ProcessQuery(r)

	requests := GetPrayerRequestsForCommunity(communityID, status, count, offset)
	Send(w, http.StatusOK, requests)
	return
}

// AddPrayerRequestToCommunityRoute adds an existing prayer request to a community
func AddPrayerRequestToCommunityRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if communityIDErr != nil || requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// first, get the request and make sure the user matches
	request, err := GetPrayerRequest(requestID)
	if err != nil || request.CreatedBy != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || (role != "admin" && role != "member") {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// alrite, add it
	err = AddPrayerRequestToCommunity(requestID, communityID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_community_add_error", "could not add that request to that community", err)
		return
	}
	Send(w, http.StatusOK, map[string]bool{
		"added": true,
	})
	return

}

// RemovePrayerRequestFromCommunityRoute removes an existing prayer request from a community
func RemovePrayerRequestFromCommunityRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	requestID, requestIDErr := strconv.ParseInt(chi.URLParam(r, "requestID"), 10, 64)
	if communityIDErr != nil || requestIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// first, get the request and make sure the user matches
	request, err := GetPrayerRequest(requestID)
	if err != nil || request.CreatedBy != jwtUser.ID {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || (role != "admin" && role != "member") {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// alrite, add it
	err = RemovePrayerRequestFromCommunity(requestID, communityID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "prayer_request_community_removed_error", "could not remove that request from that community", err)
		return
	}
	Send(w, http.StatusOK, map[string]bool{
		"removed": true,
	})
	return

}
