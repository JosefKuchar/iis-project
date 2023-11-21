package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminCategoriesRoutes() chi.Router {
	r := chi.NewRouter()

	parseForm := func(r *http.Request) (template.AdminCategoryPageData, error) {
		data := template.AdminCategoryPageData{}
		data.Errors = make(map[string]string)

		parentID, err := strconv.Atoi(r.FormValue("parent_id"))
		if err != nil {
			return data, err
		}

		idString := chi.URLParam(r, "id")
		if idString == "" {
			idString = "0"
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			return data, err
		}

		data.Category.ID = int64(id)
		data.Category.Name = r.FormValue("name")
		data.Category.ParentID = int64(parentID)
		data.Category.Approved = r.FormValue("approved") == "on"
		data.New = r.FormValue("new") == "true"

		if data.Category.Name == "" {
			data.Errors["Name"] = "Name cannot be empty"
		}

		return data, nil
	}

	getListData := func(w *http.ResponseWriter, r *http.Request) (template.AdminCategoriesPageData, error) {
		page, offset, err := getPageOffset(r)
		if err != nil {
			return template.AdminCategoriesPageData{}, err
		}

		query := r.FormValue("query")

		data := template.AdminCategoriesPageData{}
		count, err := rs.db.
			NewSelect().
			Model(&data.Categories).
			Where("name LIKE ?", "%"+query+"%").
			WhereOr("id LIKE ?", "%"+query+"%").
			Order("approved ASC").
			Limit(settings.PAGE_SIZE).
			Offset(offset).
			ScanAndCount(r.Context())
		if err != nil {
			return template.AdminCategoriesPageData{}, err
		}

		for index, category := range data.Categories {
			rs.db.NewRaw(
				`WITH RECURSIVE children as (
					SELECT c.*, 0 AS depth FROM categories c WHERE c.id = ?
					UNION ALL
					SELECT c2.*, ch.depth + 1 FROM categories as c2, children as ch
					WHERE c2.id = ch.parent_id
				)
				SELECT name, id, parent_id FROM children ORDER BY depth DESC
			`, category.ID).Scan(r.Context(), &data.Categories[index].Categories)
		}

		data.TotalCount = count
		data.Page = page
		data.Query = query

		(*w).Header().Set("HX-Push-Url", "/admin/categories?page="+strconv.Itoa(page)+"&query="+query)

		return data, nil
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		template.AdminCategoriesPage(data, appbar).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoriesPageTable(data, appbar, true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Post("/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		approved := r.FormValue("approved") == "on"
		_, err = rs.db.
			NewUpdate().
			Model(&models.Category{ID: int64(id), Approved: approved}).
			Column("approved").
			Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoriesPageTable(data, appbar, true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// New category detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminCategoryPageData{}
		data.New = true
		data.Category.Approved = true
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoryPage(data, appbar).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Create new category
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		// TODO: Check errors
		_, err = rs.db.NewInsert().Model(&data.Category).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/categories")
	})

	r.Post("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = rs.db.NewUpdate().Model(&data.Category).Where("id = ?", data.Category.ID).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/categories")
	})

	// Existing category detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminCategoryPageData{}
		data.New = false

		err := rs.db.NewSelect().Model(&data.Category).Relation("Parent").Where("category.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoryPage(data, appbar).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Delete("/{id}/table", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		_, err = rs.db.NewDelete().Model(&models.Category{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoriesPageTable(data, appbar, true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		_, err = rs.db.NewDelete().Model(&models.Category{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/categories")
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCategoryPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Get("/select2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var categories []models.Category
		err := rs.db.NewSelect().Model(&categories).Where("name LIKE ?", "%"+r.FormValue("q")+"%").Scan(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var results Select2Results
		for _, category := range categories {
			results.Results = append(results.Results, Select2Result{
				ID:   category.ID,
				Text: category.Name,
			})
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
