package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"
	"strconv"

	"JosefKuchar/iis-project/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/uptrace/bun"
)

func addTextSearch(q *bun.SelectQuery, text string) *bun.SelectQuery {
	return q.Where("event.description LIKE ?", "%"+text+"%")
}

func addUserFilter(q *bun.SelectQuery, userId interface{}) *bun.SelectQuery {
	return q.Join("JOIN user_to_event AS ute ON ute.event_id = event.id").
		Where("ute.user_id = ?", userId)
}

func addCategoryFilter(q *bun.SelectQuery, categories []int64) *bun.SelectQuery {
	return q.Join("JOIN category_to_event AS cte ON cte.event_id = event.id").
		Where("cte.category_id IN (?)", bun.In(categories))
}

func (rs resources) EventRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := template.EventsPageData{}

		rs.db.NewSelect().Model(&data.Events).Relation("Location").Relation("Categories").Scan(r.Context())
		rs.db.NewSelect().Model(&data.Categories).Scan(r.Context())

		token, _, _ := jwtauth.FromContext(r.Context())
		data.LoggedIn = token != nil

		template.EventsPage(data).Render(r.Context(), w)
	})

	r.Post("/filter", func(w http.ResponseWriter, r *http.Request) {
		var events []models.Event

		r.ParseForm()

		slug := r.FormValue("slug")
		checked := r.FormValue("myEvents")
		selectedCategories := r.Form["category"]

		q := rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories")

		if slug != "" {
			q = addTextSearch(q, slug)
		}

		if checked != "" {
			token, claims, _ := jwtauth.FromContext(r.Context())
			if token != nil {
				q = addUserFilter(q, claims["ID"])
			}
		}

		fmt.Println(selectedCategories)
		if selectedCategories != nil {
			var categories []models.Category

			rs.db.NewRaw(
				`WITH RECURSIVE cte as (
					SELECT id, name, parent_id, id as top
					FROM categories
					WHERE id IN (?)
					UNION ALL SELECT a.id, a.name, a.parent_id, b.top
					FROM categories a INNER JOIN cte b ON a.parent_id=b.id)
				SELECT id FROM cte`, bun.In(selectedCategories)).Scan(r.Context(), &categories)

			var ids []int64
			for _, item := range categories {
				ids = append(ids, item.ID)
			}

			q = addCategoryFilter(q, ids)

		}

		q = q.Order("event.id ASC").Group("event.id")

		q.Scan(r.Context())

		template.Events(events).Render(r.Context(), w)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.EventPageData{}
		id := chi.URLParam(r, "id")
		err := rs.db.
			NewSelect().
			Model(&data.Event).
			Where("event.id = ?", id).
			Relation("Location").
			Relation("Categories").
			Relation("Comments").
			Relation("Comments.User").
			Scan(r.Context())

		if err != nil {
			fmt.Println(err)
		}

		token, jwt, _ := jwtauth.FromContext(r.Context())
		if token == nil {
			data.UserId = -1
		} else {
			var u2e models.UserToEvent
			err = rs.db.NewSelect().Model(&u2e).Where("user_id = ? AND event_id = ?", jwt["ID"], id).Scan(r.Context())
			if err == nil {
				data.UserId = int(jwt["ID"].(float64))
			} else {
				data.UserId = -1
			}
		}

		data.Finished = true
		fmt.Println(data.Event.Comments)

		for _, category := range data.Event.Categories {
			var tree []models.Category
			err = rs.db.NewRaw(
				`WITH RECURSIVE children as (
					SELECT c.*, 0 AS depth FROM categories c WHERE c.id = ?
					UNION ALL
					SELECT c2.*, ch.depth + 1 FROM categories as c2, children as ch
					WHERE c2.id = ch.parent_id
				)
				SELECT name, id, parent_id FROM children ORDER BY depth DESC
				`, category.ID).Scan(r.Context(), &tree)
			data.Categories = append(data.Categories, tree)
			if err != nil {
				fmt.Println(err)
			}
		}

		template.EventPage(data).Render(r.Context(), w)
	})

	r.Post("/categories", func(w http.ResponseWriter, r *http.Request) {
		var categories []models.Category
		r.ParseForm()

		rs.db.NewSelect().Model(&categories).Where("id IN (?)", bun.In(r.Form["category"])).Scan(r.Context())

		template.Categories(categories).Render(r.Context(), w)
	})

	r.Post("/{id}/{userid}/comment", func(w http.ResponseWriter, r *http.Request) {
		commentText := r.FormValue("comment")
		eventId, _ := strconv.Atoi(chi.URLParam(r, "id"))
		userId, _ := strconv.Atoi(chi.URLParam(r, "userid"))

		comment := models.Comment{
			Text:    commentText,
			EventID: int64(eventId),
			UserID:  int64(userId),
		}

		rs.db.NewInsert().Model(&comment).Exec(r.Context())

		var comments []models.Comment

		rs.db.NewSelect().Model(&comments).Where("event_id = ?", eventId).Relation("User").Scan(r.Context())

		template.Comments(comments).Render(r.Context(), w)
	})

	return r

}
