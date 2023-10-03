package routes

import (
	"JosefKuchar/iis-project/settings"
	"net/http"
	"strconv"
)

func getPageOffset(r *http.Request) (int, int, error) {
	page := r.FormValue("page")
	if page == "" {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, 0, err
	}

	offset := (pageInt - 1) * settings.PAGE_SIZE

	return pageInt, offset, nil
}

type Select2Result struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type Select2Results struct {
	Results []Select2Result `json:"results"`
}
