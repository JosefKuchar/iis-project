package routes

import (
	"JosefKuchar/iis-project/settings"
	"net/http"
	"strconv"
)

func getOffset(r *http.Request) (int, error) {
	page := r.FormValue("page")
	if page == "" {
		page = "1"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, err
	}
	offset := (pageInt - 1) * settings.PAGE_SIZE
	return offset, nil
}
