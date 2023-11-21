package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/template"
	"net/http"
)

func getAppbarData(rs *resources, r *http.Request) (template.AppbarData, error) {
	data := template.AppbarData{}
	data.User = models.User{}
	data.NewCategories = 6

	return data, nil
}
