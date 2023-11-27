package routes

import (
	"JosefKuchar/iis-project/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"JosefKuchar/iis-project/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/uptrace/bun"
)

func calculateAverageRating(ratings []models.Rating) float64 {
	var sum int64
	for _, rating := range ratings {
		sum += rating.Rating
	}
	var len = len(ratings)
	if len == 0 {
		return 0.0
	} else {
		return float64(sum) / float64(len)
	}
}

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

		err := rs.db.NewSelect().Model(&data.Events).Where("event.approved = 1").Relation("Location").Relation("Categories").Relation("Ratings").Scan(r.Context())
		if err != nil {
			fmt.Println(err)
		}

		err = rs.db.NewSelect().Model(&data.Categories).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
		}

		err = rs.db.NewSelect().Model(&data.Locations).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
		}

		token, _, _ := jwtauth.FromContext(r.Context())
		data.LoggedIn = token != nil

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data.AverageRatings = make([]float64, len(data.Events))

		for i, event := range data.Events {
			data.AverageRatings[i] = calculateAverageRating(event.Ratings)
		}

		err = template.EventsPage(data, appbar).Render(r.Context(), w)
		if err != nil {
			fmt.Println(err)
		}
	})

	r.Post("/filter", func(w http.ResponseWriter, r *http.Request) {
		var events []models.Event

		r.ParseForm()

		slug := r.FormValue("slug")
		checked := r.FormValue("myEvents")
		from := r.FormValue("from")
		to := r.FormValue("to")
		location := r.FormValue("location")
		selectedCategories := r.Form["categories"]

		q := rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories").Relation("Ratings")

		if slug != "" {
			q = addTextSearch(q, slug)
		}

		fmt.Println(location)
		if location != "" {
			q = q.Where("location.id = ?", location)
		}

		if from != "" {
			q = q.Where("event.start >= ?", from)
		}

		if to != "" {
			q = q.Where("event.end <= ?", to)
		}

		if checked != "" {
			token, claims, _ := jwtauth.FromContext(r.Context())
			if token != nil {
				q = addUserFilter(q, claims["ID"])
			}
		}

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

		var ratings = make([]float64, len(events))

		for i, event := range events {
			ratings[i] = calculateAverageRating(event.Ratings)
		}

		template.Events(events, ratings).Render(r.Context(), w)
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
			Relation("UsersToEvent").
			Relation("Comments").
			Relation("Ratings").
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

		// Check capacity
		userToEventCount, err := rs.db.NewSelect().
			Model((*models.UserToEvent)(nil)).
			Where("user_to_event.event_id = ?", id).
			Count(r.Context())

		data.Full = userToEventCount >= int(data.Event.Capacity)

		var userRating models.Rating

		err = rs.db.NewSelect().Model(&userRating).Where("user_id = ? AND event_id = ?", data.UserId, id).Scan(r.Context())

		if err != nil {
			data.UserRating = 0
		} else {
			data.UserRating = userRating.Rating
		}

		data.AverageRating = calculateAverageRating(data.Event.Ratings)

		fmt.Println(data.AverageRating)

		template.EventPage(data, appbar).Render(r.Context(), w)
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

		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		template.Comments(comments, appbar).Render(r.Context(), w)
	})

	r.Post("/{id}/{userid}/rate", func(w http.ResponseWriter, r *http.Request) {
		eventId, _ := strconv.Atoi(chi.URLParam(r, "id"))
		userId, _ := strconv.Atoi(chi.URLParam(r, "userid"))

		var rating, _ = strconv.Atoi(r.FormValue("rating-9"))

		rs.db.NewDelete().Model(&models.Rating{}).Where("user_id = ? AND event_id = ?", userId, eventId).Exec(r.Context())
		if rating != 0 {
			userRating := models.Rating{
				UserID:  int64(userId),
				EventID: int64(eventId),
				Rating:  int64(rating),
			}
			rs.db.NewInsert().Model(&userRating).Exec(r.Context())
		}

		var ratings []models.Rating
		rs.db.NewSelect().Model(&ratings).Where("event_id = ?", eventId).Scan(r.Context())

		var averageRating = calculateAverageRating(ratings)

		template.AverageRating(averageRating, eventId).Render(r.Context(), w)

	})

	r.Get("/categories/select2", func(w http.ResponseWriter, r *http.Request) {
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

	r.Get("/locations/select2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var locations []models.Location
		err := rs.db.NewSelect().Model(&locations).Where("name LIKE ?", "%"+r.FormValue("q")+"%").Scan(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var results Select2Results
		for _, location := range locations {
			results.Results = append(results.Results, Select2Result{
				ID:   location.ID,
				Text: location.Name,
			})
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
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

			template.RegisterSection(userId, false, eventId, entranceFees, nil, false).Render(r.Context(), w)
			return
		}

		// Check if we reached the capacity
		var event models.Event
		err = rs.db.NewSelect().
			Model(&event).
			Where("event.id = ?", eventId).
			Relation("EntranceFees").
			Scan(r.Context())

		if err != nil {
			fmt.Println(err)
		}

		userToEventCount, err := rs.db.NewSelect().
			Model((*models.UserToEvent)(nil)).
			Where("user_to_event.event_id = ?", eventId).
			Count(r.Context())

		if err != nil {
			fmt.Println(err)
		}

		eventFull := userToEventCount >= int(event.Capacity)

		template.RegisterSection(userId, true, eventId, nil, &userToEvent, eventFull).Render(r.Context(), w)

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
