package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/template"
	"net/http"
)

func getAppbarData(rs *resources, r *http.Request) (template.AppbarData, error) {
	data := template.AppbarData{}
	data.User = models.User{}

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

	return data, nil
}
