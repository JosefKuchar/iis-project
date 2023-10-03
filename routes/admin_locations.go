package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminLocationsRoutes() chi.Router {
	r := chi.NewRouter()

	parseForm := func(r *http.Request) template.AdminLocationPageData {
		data := template.AdminLocationPageData{}
		data.Errors = make(map[string]string)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			fmt.Println(err)
		}

		data.Location.ID = int64(id)
		data.Location.Name = r.FormValue("name")
		data.Location.Approved = r.FormValue("approved") == "on"
		data.New = r.FormValue("new") == "true"

		if data.Location.Name == "" {
			data.Errors["Name"] = "Name cannot be empty"
		}

		return data
	}

	getListData := func(w *http.ResponseWriter, r *http.Request) (template.AdminLocationsPageData, error) {
		page, offset, err := getPageOffset(r)
		query := r.FormValue("query")
		if err != nil {
			return template.AdminLocationsPageData{}, err
		}
		data := template.AdminLocationsPageData{}
		count, err := rs.db.
			NewSelect().
			Model(&data.Locations).
			Where("name LIKE ?", "%"+query+"%").
			WhereOr("id LIKE ?", "%"+query+"%").
			Order("approved ASC").
			Limit(settings.PAGE_SIZE).
			Offset(offset).
			ScanAndCount(r.Context())
		if err != nil {
			return template.AdminLocationsPageData{}, err
		}

		data.TotalCount = count
		data.Page = page
		data.Query = query
		(*w).Header().Set("HX-Push-Url", "/admin/locations?page="+strconv.Itoa(page)+"&query="+query)
		return data, nil
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		template.AdminLocationsPage(data).Render(r.Context(), w)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		template.AdminLocationsPageTable(data).Render(r.Context(), w)
	})

	r.Post("/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			fmt.Println(err)
		}
		approved := r.FormValue("approved") == "on"
		_, err = rs.db.
			NewUpdate().
			Model(&models.Location{ID: int64(id), Approved: approved}).
			Column("approved").
			Where("id = ?", id).Exec(r.Context())
		if err != nil {
			fmt.Println(err)
		}
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		template.AdminLocationsPageTable(data).Render(r.Context(), w)
	})

	// New location detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminLocationPageData{}
		data.New = true
		template.AdminLocationPage(data).Render(r.Context(), w)
	})

	// Create new location
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		data := parseForm(r)

		// TODO: Check errors
		// Create new location
		rs.db.NewInsert().Model(&data.Location).Exec(r.Context())
		w.Header().Set("HX-Redirect", "/admin/locations")
	})

	r.Post("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := parseForm(r)

		rs.db.NewUpdate().Model(&data.Location).Where("id = ?", data.Location.ID).Exec(r.Context())
		w.Header().Set("HX-Redirect", "/admin/locations")
	})

	// Existing location detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminLocationPageData{}
		data.New = false
		err := rs.db.NewSelect().Model(&data.Location).Where("location.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
		template.AdminLocationPage(data).Render(r.Context(), w)
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		data := parseForm(r)

		err := template.AdminLocationPageForm(data).Render(r.Context(), w)
		if err != nil {
			fmt.Println(err)
		}
	})

	return r
}
