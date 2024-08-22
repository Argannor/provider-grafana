package common

import (
	"crypto/rand"
	"net/http"
	"strconv"

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"

	grafana "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/pkg/errors"
)

// we ignore forbidden messages on observations, as we cannot discern between
// - organization (containing the resource) is missing
// - user does not have access to the organization
// as both result in a 403 response.
// 404 is returned iff the user has access to the organization and the resource type, but the resource is missing
var ignoreStatusCodesOnObserve = []int{http.StatusForbidden, http.StatusNotFound}

type ApiError interface {
	error
	IsCode(code int) bool
}

type ApiResponse[R interface{}] interface {
	IsCode(code int) bool
	GetPayload() *R
}

type GrafanaAPI struct {
	service grafana.GrafanaHTTPAPI
}

func NewGrafanaAPI(service grafana.GrafanaHTTPAPI) GrafanaAPI {
	return GrafanaAPI{service: service}
}

func (g *GrafanaAPI) GetAllUsers() ([]*models.UserSearchHitDTO, error) {
	var allUsers []*models.UserSearchHitDTO
	var page int64 = 0
	params := users.NewSearchUsersParams().WithDefaults()
	client := g.service.Clone()

	for {
		resp, err := client.Users.SearchUsers(params.WithPage(&page), nil)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, resp.Payload...)
		if len(resp.Payload) != int(*params.Perpage) {
			break
		}
		page++
	}
	return allUsers, nil
}

func (g *GrafanaAPI) CreateUser(user string) (int64, error) {
	client := g.service.Clone()
	n := 64
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0, err
	}
	pass := string(bytes[:n])
	u := models.AdminCreateUserForm{
		Name:     user,
		Login:    user,
		Email:    user,
		Password: pass,
	}
	resp, err := client.AdminUsers.AdminCreateUser(&u)
	if err != nil {
		return 0, err
	}
	return resp.Payload.ID, err
}

func (g *GrafanaAPI) GetAllOrgs() ([]*models.OrgDTO, error) {
	var allOrgs []*models.OrgDTO
	var page int64 = 0
	params := orgs.NewSearchOrgsParams().WithDefaults()
	client := g.service
	for {
		resp, err := client.Orgs.SearchOrgs(params.WithPage(&page), nil)
		if err != nil {
			return nil, err
		}

		allOrgs = append(allOrgs, resp.Payload...)
		if len(resp.Payload) != int(*params.Perpage) {
			break
		}
		page++
	}
	return allOrgs, nil
}

// SwitchToLowestOrgId switches the current user's active organization to the one with the lowest ID.
// It first retrieves all organizations and iterates through them to find the one with the lowest ID.
// Then, it uses the Grafana API to switch the current user's active organization to the one found.
// This function is useful in scenarios where the user needs to be in a context that is not the organization being managed,
// for example, when deleting an organization.
//
// Returns:
//
//	error: If an error occurred during the process. It could be due to issues in retrieving all organizations or switching the active organization.
func (g *GrafanaAPI) SwitchToLowestOrgId() error {
	orgas, err := g.GetAllOrgs()
	if err != nil {
		return err
	}
	var orgId int64
	orgId = 9999999
	for _, org := range orgas {
		if org.ID < orgId {
			orgId = org.ID
		}
	}
	_, err = g.service.SignedInUser.UserSetUsingOrg(orgId)
	return err
}

func (g *GrafanaAPI) GetSignedInUser() (*models.UserProfileDTO, error) {
	resp, err := g.service.SignedInUser.GetSignedInUser()
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) UserSetUsingOrg(orgId int64) (*models.SuccessResponseBody, error) {
	resp, err := g.service.Clone().WithOrgID(0).SignedInUser.UserSetUsingOrg(orgId)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) CreateOrg(name string) (*models.CreateOrgOKBody, error) {
	cmd := &models.CreateOrgCommand{
		Name: name,
	}
	resp, err := g.service.Clone().WithOrgID(0).Orgs.CreateOrg(cmd)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) DeleteOrgByID(orgID int64) (*models.SuccessResponseBody, error) {
	resp, err := g.service.WithOrgID(0).Orgs.DeleteOrgByID(orgID)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) AddOrgUser(orgID int64, user *models.AddOrgUserCommand) (*models.SuccessResponseBody, error) {
	resp, err := g.service.Orgs.AddOrgUser(orgID, user)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) UpdateOrgUser(orgID int64, userID int64, user *models.UpdateOrgUserCommand) (*models.SuccessResponseBody, error) {
	params := orgs.NewUpdateOrgUserParams().WithOrgID(orgID).WithUserID(userID).WithBody(user)
	resp, err := g.service.Orgs.UpdateOrgUser(params)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) RemoveOrgUser(userID int64, orgID int64) (*models.SuccessResponseBody, error) {
	resp, err := g.service.Orgs.RemoveOrgUser(userID, orgID)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) AdminCreateUser(user *models.AdminCreateUserForm) (*models.AdminCreateUserResponse, error) {
	resp, err := g.service.AdminUsers.AdminCreateUser(user)
	if err != nil {
		return nil, err
	}
	return resp.Payload, err
}

func (g *GrafanaAPI) GetOrgByName(s string) (*models.OrgDetailsDTO, error) {
	response, err := g.service.Orgs.GetOrgByName(s)
	return orNilOnNotFound[models.OrgDetailsDTO](&response, err)
}

func (g *GrafanaAPI) GetOrgById(id int64) (*models.OrgDetailsDTO, error) {
	response, err := g.service.Orgs.GetOrgByID(id)
	return orNilOnNotFound[models.OrgDetailsDTO](&response, err)
}

func (g *GrafanaAPI) GetOrgUsers(orgId int64) ([]*models.OrgUserDTO, error) {
	response, err := g.service.Orgs.GetOrgUsers(orgId)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) GetDataSourceById(orgId int64, id string) (*models.DataSource, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Datasources.GetDataSourceByID(id)
	return orNilOnStatus[models.DataSource](&response, err, ignoreStatusCodesOnObserve...)
}

func (g *GrafanaAPI) GetDataSourceByName(orgId int64, name string) (*models.DataSource, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Datasources.GetDataSourceByName(name)
	return orNilOnStatus[models.DataSource](&response, err, ignoreStatusCodesOnObserve...)
}

func (g *GrafanaAPI) CreateDataSource(orgId int64, command *models.AddDataSourceCommand) (*models.AddDataSourceOKBody, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Datasources.AddDataSource(command)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) UpdateDataSource(orgId int64, id string, command *models.UpdateDataSourceCommand) (*models.UpdateDataSourceByIDOKBody, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Datasources.UpdateDataSourceByID(id, command)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) DeleteDataSource(orgId int64, id string) (*models.SuccessResponseBody, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Datasources.DeleteDataSourceByID(id)
	if err != nil {
		return nil, err

	}
	return response.Payload, err
}

func (g *GrafanaAPI) CreateOrUpdateDashboard(orgId int64, command *models.SaveDashboardCommand) (*models.PostDashboardOKBody, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Dashboards.PostDashboard(command)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) GetDashboardByUid(orgId int64, uid string) (*models.DashboardFullWithMeta, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Dashboards.GetDashboardByUID(uid)
	return orNilOnStatus[models.DashboardFullWithMeta](&response, err, ignoreStatusCodesOnObserve...)
}

func (g *GrafanaAPI) GetDashboardByName(orgId int64, name string, folder *string) (*models.DashboardFullWithMeta, error) {
	dashboardType := "dash-db"
	params := &search.SearchParams{
		Type:  &dashboardType,
		Query: &name,
	}
	setFolderIdIfNotNull(folder, params)
	response, err := g.service.Clone().WithOrgID(orgId).Search.Search(params)
	if err != nil {
		return nil, err
	}
	if len(response.Payload) == 0 {
		return nil, nil
	}
	uid := response.Payload[0].UID
	return g.GetDashboardByUid(orgId, uid)
}

func setFolderIdIfNotNull(folder *string, params *search.SearchParams) {
	if folder == nil {
		return
	}
	folderId, err := strconv.ParseInt(*folder, 10, 64)
	if err == nil {
		params.FolderIds = []int64{folderId}
	} else {
		params.FolderUIDs = []string{*folder}
	}
}

func (g *GrafanaAPI) DeleteDashboard(orgId int64, uid string) (*models.DeleteDashboardByUIDOKBody, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Dashboards.DeleteDashboardByUID(uid)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) GetFolderByUid(orgId int64, uid string) (*models.Folder, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Folders.GetFolderByUID(uid)
	return orNilOnStatus[models.Folder](&response, err, ignoreStatusCodesOnObserve...)
}

func (g *GrafanaAPI) GetFolderById(orgId int64, id int64) (*models.Folder, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Folders.GetFolderByID(id)
	return orNilOnStatus[models.Folder](&response, err, ignoreStatusCodesOnObserve...)
}

func (g *GrafanaAPI) GetFolderByName(orgId int64, name string, parentFolder *string) (*models.Folder, error) {
	dashboardType := "dash-folder"
	params := &search.SearchParams{
		Type:  &dashboardType,
		Query: &name,
	}
	setFolderIdIfNotNull(parentFolder, params)
	response, err := g.service.Clone().WithOrgID(orgId).Search.Search(params)
	if err != nil {
		return nil, err
	}
	if len(response.Payload) == 0 {
		return nil, nil
	}
	uid := response.Payload[0].UID
	return g.GetFolderByUid(orgId, uid)
}

func (g *GrafanaAPI) CreateFolder(orgId int64, command *models.CreateFolderCommand) (*models.Folder, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Folders.CreateFolder(command)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) UpdateFolder(orgId int64, uid string, command *models.UpdateFolderCommand) (*models.Folder, error) {
	response, err := g.service.Clone().WithOrgID(orgId).Folders.UpdateFolder(uid, command)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func (g *GrafanaAPI) DeleteFolder(orgId int64, uid string) (*models.DeleteFolderOKBody, error) {
	deleteRules := false
	params := folders.DeleteFolderParams{
		FolderUID:        uid,
		ForceDeleteRules: &deleteRules,
	}
	response, err := g.service.Clone().WithOrgID(orgId).Folders.DeleteFolder(&params)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func orNilOnNotFound[R interface{}, T ApiResponse[R]](response *T, err error) (*R, error) {
	return orNilOnStatus[R, T](response, err, 404)
}

func orNilOnStatus[R interface{}, T ApiResponse[R]](response *T, err error, status ...int) (*R, error) {
	if err != nil && isCode(err, status...) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	for _, code := range status {
		if (*response).IsCode(code) {
			return nil, nil
		}
	}
	return (*response).GetPayload(), err
}

// nolint: unparam
func isCode(err error, codes ...int) bool {
	if err == nil {
		return false
	}
	var oasError ApiError
	isOasError := errors.As(err, &oasError)
	if isOasError {
		for _, code := range codes {
			if oasError.IsCode(code) {
				return true
			}
		}
	}
	return false
}
