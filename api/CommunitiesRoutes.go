package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Bind binds data
func (data *Community) Bind(r *http.Request) error {
	return nil
}

// Bind binds data
func (data *CommunityUserLink) Bind(r *http.Request) error {
	return nil
}

// CreateCommunityRoute creates a community
func CreateCommunityRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	input := Community{}
	render.Bind(r, &input)
	input.Name, _ = sanitize(input.Name)
	input.Description, _ = sanitize(input.Description)

	if input.Name == "" {
		SendError(w, http.StatusBadRequest, "community_create_error", "name is required", nil)
		return
	}

	if input.ShortCode == "" {
		rand.Seed(time.Now().UnixNano())
		nameLength := len(input.Name)
		nameCode := input.Name
		if nameLength > 10 {
			nameCode = input.Name[:11]
		}
		input.ShortCode = fmt.Sprintf("%s%d%d%d%d", strings.ToLower(nameCode), rand.Int63n(9), rand.Int63n(9), rand.Int63n(9), jwtUser.ID)
	}

	if input.Privacy == "" {
		input.Privacy = CommunityPrivacyPrivate
	}

	if input.UserSignupStatus == "" {
		input.UserSignupStatus = CommunityUserSignupStatusApproval
	}

	if input.Plan == "" {
		input.Plan = CommunityPlanFree
	}

	input.PlanPaidThrough = time.Now().Format("2006-01-02")
	input.PlanDiscountPercent = 0

	// verify the name doesn't already exist and isn't a
	// reserved community name
	foundByName, err := GetCommunityByName(input.Name)
	if err == nil && foundByName.Name == input.Name {
		// it already exists
		SendError(w, http.StatusConflict, "community_create_exists", "that name is taken or reserved", input)
		return
	}

	// ditto on the short code
	foundByCode, err := GetCommunityByShortCode(input.ShortCode)
	if err == nil && foundByCode.ShortCode == input.ShortCode {
		// it already exists
		SendError(w, http.StatusConflict, "community_create_exists", "that short code is taken or reserved", input)
		return
	}

	err = CreateCommunity(&input)
	if err != nil {
		SendError(w, 400, "community_create_error", "could not create the community", err)
		return
	}

	// created, join the current user as the owner
	err = CreateCommunityUserLink(input.ID, jwtUser.ID, "admin", "accepted", "")
	if err != nil {
		SendError(w, 400, "community_create_error", "could not add the user to the community", err)
		return
	}

	Send(w, 201, input)
	return
}

// UpdateCommunityRoute updates some of the basic information about a community
func UpdateCommunityRoute(w http.ResponseWriter, r *http.Request) {
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

	input := Community{}
	render.Bind(r, &input)
	input.Name, _ = sanitize(input.Name)
	input.Description, _ = sanitize(input.Description)

	community, err := GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || role != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	if input.Name != "" {
		// verify the name doesn't already exist and isn't a
		// reserved community name
		foundByName, err := GetCommunityByName(input.Name)
		if err == nil && foundByName.Name == input.Name {
			// it already exists
			SendError(w, http.StatusConflict, "community_create_name_exists", "that name is taken or reserved", input)
			return
		}
		community.Name = input.Name
	}

	if input.Description != "" {
		community.Description = input.Description
	}

	if input.JoinCode != "" {
		community.JoinCode = input.JoinCode
	}

	if input.Privacy != "" {
		community.Privacy = input.Privacy
	}

	if input.UserSignupStatus != "" {
		community.UserSignupStatus = input.UserSignupStatus
	}

	err = UpdateCommunity(community)
	if err != nil {
		SendError(w, http.StatusForbidden, "community_update_error", "could not update that community", err)
		return
	}
	Send(w, http.StatusOK, community)
	return
}

// DeleteCommunityRoute deletes a community, links to the prayers, and any users on the group
func DeleteCommunityRoute(w http.ResponseWriter, r *http.Request) {
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

	_, err = GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || role != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	err = DeleteCommunity(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "community_delete_error", "could not delete that community", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
	return

}

// GetCommunityByIDRoute gets a single community
func GetCommunityByIDRoute(w http.ResponseWriter, r *http.Request) {
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

	community, err := GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	role, joinErr := GetUserRoleForCommunity(communityID, jwtUser.ID)

	if community.Privacy == "private" && (role == "" || joinErr != nil) {
		SendError(w, http.StatusForbidden, "community_user_not_member", "you are not a member of this community", err)
		return
	}

	// if the user is not an admin, we strip out some field
	if role != "admin" {
		community.JoinCode = ""
		community.ShortCode = ""
		community.UserSignupStatus = ""
		community.PlanPaidThrough = ""
		community.PlanDiscountPercent = 0
		community.StripeSubscriptionID = ""
	}

	Send(w, http.StatusOK, community)
	return

}

// GetCommunitiesForUserRoute gets the communities the user is attached to
func GetCommunitiesForUserRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communities, _ := GetCommunitiesForUser(jwtUser.ID)
	Send(w, http.StatusOK, communities)
	return
}

// GetPublicCommunitiesRoute gets the public, non-site communities
func GetPublicCommunitiesRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}
	_, _, count, offset, sortField, sortDir, _, _ := ProcessQuery(r)
	communities, _ := GetPublicCommunities(sortField, sortDir, count, offset)
	Send(w, http.StatusOK, communities)
	return
}

// RequestCommunityMembershipRoute allows a user to request membership in a public community
func RequestCommunityMembershipRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// there are two distinct paths here
	// one is that the user is requesting that they can join a public community; the userId and the jwt user id will be the same
	// otherwise, the user is requesting another user join a community they are an admin for

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if communityIDErr != nil || userIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	community, err := GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	// figure out the plan status and number of users allowed
	plan := plans[community.Plan]
	currentCount, _ := GetCountOfUsersInCommunity(community.ID)
	if plan.AllowedUsers <= currentCount {
		SendError(w, http.StatusForbidden, "membership_full", "this community cannot accept anymore members", map[string]interface{}{
			"currentCount": currentCount,
			"allowed":      plan.AllowedUsers,
		})
		return
	}

	// now we break off
	if userID == jwtUser.ID {
		// the user is requesting that they join the community
		if community.Privacy == "private" {
			SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
			return
		}
		// if the community auto accepts, just add them
		if community.UserSignupStatus == CommunityUserSignupStatusAccept {
			err = CreateCommunityUserLink(community.ID, jwtUser.ID, "member", "accepted", "")
			Send(w, http.StatusOK, map[string]bool{
				"joined": true,
			})
		}
		// generate a shortCode and create the request
		code := GenerateShortCode(communityID, jwtUser.ID)

		err = CreateCommunityUserLink(community.ID, jwtUser.ID, "member", "requested", code)
		if err != nil {
			SendError(w, http.StatusBadRequest, "membership_request_error", "could not request membership", err)
			return
		}

		// TODO: send an email to the admin of the community

		Send(w, http.StatusOK, map[string]bool{
			"requested": true,
		})
		return
	}

	// no, the jwtUser is requesting another user join the community, so we need to check some permissions
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || role != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// ok, now we create the link
	// generate a shortCode and create the request
	code := GenerateShortCode(communityID, userID)

	err = CreateCommunityUserLink(community.ID, userID, "member", "invited", code)
	if err != nil {
		SendError(w, http.StatusBadRequest, "membership_request_error", "could not request user joins", err)
		return
	}

	// TODO: send an email to the user asking them if they want to join

	Send(w, http.StatusOK, map[string]bool{
		"invited": true,
	})
	return

}

// RemoveCommunityMembershipRoute removes a user from a community, whether accepted or not
func RemoveCommunityMembershipRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if communityIDErr != nil || userIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	_, err = GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	// this endpoint is solely for admins
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || role != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	err = DeleteCommunityUserLink(communityID, userID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "community_user_link_error", "could not delete that link", err)
		return
	}

	Send(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
	return
}

// ProcessCommunityMembershipRoute updates the link for a request
func ProcessCommunityMembershipRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if communityIDErr != nil || userIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	// similar to requesting, there are two branches here
	// one for an admin approving a request and one for the user approving an invitation

	input := CommunityUserLink{}
	render.Bind(r, &input)
	if input.ShortCode == "" && input.Status == "" {
		SendError(w, http.StatusBadRequest, "community_user_link_error", "shortCode and status is required", input)
		return
	}

	// get the link
	link, err := GetCommunityUserLink(communityID, userID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	if link.Status == "accepted" {
		SendError(w, http.StatusBadRequest, "community_user_link_already_accepted", "link is already accepted and cannot be modified here", link)
		return
	}

	if link.ShortCode != input.ShortCode {
		SendError(w, http.StatusForbidden, "community_user_link_error", "shortCode does not match", input)
		return
	}

	// so, which branch is it?
	if userID == jwtUser.ID {
		// it is a user managing their own invitation
		if link.Status != "invited" {
			SendError(w, http.StatusBadRequest, "community_user_link_not_invited", "you can only manage invitations where the status is not invited", link)
			return
		}
		// update the link
		err = UpdateCommunityUserLink(communityID, userID, input.Status)
		if err != nil {
			SendError(w, http.StatusForbidden, "community_user_link_error", "could not update the link", err)
			return
		}

		link, _ = GetCommunityUserLink(communityID, userID)
		Send(w, http.StatusOK, link)
		return
	}

	// it is an admin approving a request
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || role != "admin" {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	if link.Status != "requested" {
		SendError(w, http.StatusBadRequest, "community_user_link_not_requested", "the status must be 'requested'", link)
		return
	}

	// update the link
	err = UpdateCommunityUserLink(communityID, userID, input.Status)
	if err != nil {
		SendError(w, http.StatusForbidden, "community_user_link_error", "could not update the link", err)
		return
	}

	link, _ = GetCommunityUserLink(communityID, userID)
	Send(w, http.StatusOK, link)
	return
}

// GetCommunityLinksRoute gets the links for a community
func GetCommunityLinksRoute(w http.ResponseWriter, r *http.Request) {
	jwtUser, err := CheckForUser(r)
	if err != nil || jwtUser.ID == 0 {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "all"
	}

	communityID, communityIDErr := strconv.ParseInt(chi.URLParam(r, "communityID"), 10, 64)
	if communityIDErr != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", nil)
		return
	}

	_, err = GetCommunityByID(communityID)
	if err != nil {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	// in order to see this list, you need to either be an admin or member
	role, err := GetUserRoleForCommunity(communityID, jwtUser.ID)
	if err != nil || (role != "admin" && role != "member") {
		SendError(w, http.StatusForbidden, "permission_denied", "you don't have permission", err)
		return
	}

	links, _ := GetCommunityUserLinks(communityID, status)
	// if they are an admin, remove all of the short codes
	if role != "admin" {
		for i := range links {
			links[i].ShortCode = ""
		}
	}

	Send(w, http.StatusOK, links)
	return
}
