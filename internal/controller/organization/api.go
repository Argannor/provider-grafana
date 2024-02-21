package organization

import (
	"crypto/rand"
	grafana "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/pkg/errors"
)

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
	orgs, err := g.GetAllOrgs()
	if err != nil {
		return err
	}
	var orgId int64
	orgId = 9999999
	for _, org := range orgs {
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
	org, err := g.service.Orgs.GetOrgByName(s)
	if err != nil {
		if isCode(err, 404) {
			return nil, nil
		}
		return nil, err
	}
	return org.Payload, nil

}

func (g *GrafanaAPI) GetOrgUsers(orgId int64) ([]*models.OrgUserDTO, error) {
	response, err := g.service.Orgs.GetOrgUsers(orgId)
	if err != nil {
		return nil, err
	}
	return response.Payload, err
}

func isCode(err error, code int) bool {
	if err == nil {
		return false
	}
	var oasError ApiError
	isOasError := errors.As(err, &oasError)
	if isOasError {
		return oasError.IsCode(code)
	}
	return false
}
