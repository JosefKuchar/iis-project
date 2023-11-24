package routes

import (
	"JosefKuchar/iis-project/models"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"JosefKuchar/iis-project/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/uptrace/bun"
)

func addTextSearch(q *bun.SelectQuery, text string) *bun.SelectQuery {

	// TODO: maybe add full text for category filtering as well
	q = q.WhereGroup("OR", func(q2 *bun.SelectQuery) *bun.SelectQuery {
		return q2.Where("event.name LIKE ?", "%"+text+"%").WhereOr("location.name LIKE ?", "%"+text+"%").WhereOr("event.description LIKE ?", "%"+text+"%")
	})

	return q
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

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		template.EventsPage(data, appbar).Render(r.Context(), w)
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

		fmt.Println(q)

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
			Relation("EntranceFees").
			Scan(r.Context())

		if err != nil {
			fmt.Println(err)
		}

		token, jwt, _ := jwtauth.FromContext(r.Context())
		if token == nil {
			data.UserId = -1
		} else {
			data.UserId = int(jwt["ID"].(float64))

			var u2e models.UserToEvent
			err = rs.db.NewSelect().
				Model(&u2e).
				Where("user_to_event.user_id = ? AND user_to_event.event_id = ?", jwt["ID"], id).
				Relation("EntranceFee").
				Scan(r.Context())
			if err == nil {
				data.RegisteredFee = &u2e
			}
		}

		data.Finished = data.Event.End.Before(time.Now())
		fmt.Println(data.Finished)

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

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		template.EventPage(data, appbar).Render(r.Context(), w)
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

	r.With(validateAction).Post("/{id}/{userid}/{entrancefeeid}/{action}", func(w http.ResponseWriter, r *http.Request) {
		eventId, _ := strconv.Atoi(chi.URLParam(r, "id"))
		userId, _ := strconv.Atoi(chi.URLParam(r, "userid"))
		entranceFeeId, _ := strconv.Atoi(chi.URLParam(r, "entrancefeeid"))
		action := chi.URLParam(r, "action")

		if action == "register" {
			userToEvent := models.UserToEvent{
				UserID:        int64(userId),
				EventID:       int64(eventId),
				EntranceFeeID: int64(entranceFeeId),
			}
			rs.db.NewInsert().Model(&userToEvent).Exec(r.Context())
		} else {
			rs.db.NewDelete().Model(&models.UserToEvent{}).Where("user_id = ? AND event_id = ?", userId, eventId).Exec(r.Context())
		}

		// Get the userToEvent
		var userToEvent models.UserToEvent
		err := rs.db.NewSelect().
			Model(&userToEvent).
			Where("user_to_event.user_id = ?", userId).
			Where("user_to_event.event_id = ?", eventId).
			Relation("EntranceFee").
			Scan(r.Context())

		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		}

		if err == sql.ErrNoRows {
			// User is no longer registered, get entrance fees
			var entranceFees []models.EntranceFee
			err = rs.db.NewSelect().
				Model(&entranceFees).
				Where("entrance_fee.event_id = ?", eventId).
				Scan(r.Context())

			template.RegisterSection(userId, false, eventId, entranceFees, nil).Render(r.Context(), w)
			return
		}

		template.RegisterSection(userId, true, eventId, nil, &userToEvent).Render(r.Context(), w)

	})

	return r

}

func validateAction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := chi.URLParam(r, "action")
		if action != "register" && action != "unregister" {
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
