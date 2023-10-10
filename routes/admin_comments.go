package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminCommentsRoutes() chi.Router {
	r := chi.NewRouter()

	parseForm := func(r *http.Request) (template.AdminCommentPageData, error) {
		data := template.AdminCommentPageData{}
		data.Errors = make(map[string]string)

		idString := chi.URLParam(r, "id")
		if idString == "" {
			idString = "0"
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			return data, err
		}

		data.Comment.ID = int64(id)
		data.Comment.Text = r.FormValue("text")

		if data.Comment.Text == "" {
			data.Errors["Text"] = "Text cannot be empty"
		}

		return data, nil
	}

	getListData := func(w *http.ResponseWriter, r *http.Request) (template.AdminCommentsPageData, error) {
		page, offset, err := getPageOffset(r)
		if err != nil {
			return template.AdminCommentsPageData{}, err
		}

		query := r.FormValue("query")

		data := template.AdminCommentsPageData{}
		count, err := rs.db.
			NewSelect().
			Model(&data.Comments).
			Relation("User").
			Relation("Event").
			Where("text LIKE ?", "%"+query+"%").
			WhereOr("comment.id LIKE ?", "%"+query+"%").
			WhereOr("user.name LIKE ?", "%"+query+"%").
			WhereOr("event.name LIKE ?", "%"+query+"%").
			Limit(settings.PAGE_SIZE).
			Offset(offset).
			ScanAndCount(r.Context())
		if err != nil {
			return template.AdminCommentsPageData{}, err
		}

		data.TotalCount = count
		data.Page = page
		data.Query = query

		(*w).Header().Set("HX-Push-Url", "/admin/comments?page="+strconv.Itoa(page)+"&query="+query)

		return data, nil
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCommentsPage(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCommentsPageTable(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Post("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = rs.db.NewUpdate().Model(&data.Comment).Where("id = ?", data.Comment.ID).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/comments")
	})

	// Existing comment detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminCommentPageData{}

		err := rs.db.NewSelect().Model(&data.Comment).Where("comment.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		err = template.AdminCommentPage(data).Render(r.Context(), w)
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

		_, err = rs.db.NewDelete().Model(&models.Comment{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCommentsPageTable(data).Render(r.Context(), w)
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

		_, err = rs.db.NewDelete().Model(&models.Comment{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/comments")
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminCommentPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
