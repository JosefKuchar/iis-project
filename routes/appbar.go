package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func getAppbarData(rs *resources, r *http.Request) (template.AppbarData, error) {
	data := template.AppbarData{}
	data.User = template.UserData{}

	// Fetch number of new events
	count, err := rs.db.NewSelect().Model((*models.Event)(nil)).Where("approved = false").Count(r.Context())
	if err != nil {
		return data, err
	}
	data.NewEvents = count
	// Fetch number of new locations
	count, err = rs.db.NewSelect().Model((*models.Location)(nil)).Where("approved = false").Count(r.Context())
	if err != nil {
		return data, err
	}
	data.NewLocations = count
	// Fetch number of new categories
	count, err = rs.db.NewSelect().Model((*models.Category)(nil)).Where("approved = false").Count(r.Context())
	if err != nil {
		return data, err
	}
	data.NewCategories = count

	// Get user data from token
	token, claims, _ := jwtauth.FromContext(r.Context())
	if token != nil {
		data.User.ID = int(claims["ID"].(float64))
		data.User.Name = claims["Name"].(string)
		data.User.Email = claims["Email"].(string)
		data.User.RoleID = int(claims["RoleID"].(float64))
		data.User.RoleName = (claims["Role"].(map[string]interface{}))["Name"].(string)
	}

	return data, nil
}
