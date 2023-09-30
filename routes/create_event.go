package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (rs resources) CreateEventRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := template.CreateEventData{}

		rs.db.NewSelect().Model(&data.Categories).Scan(r.Context())
		rs.db.NewSelect().Model(&data.Locations).Scan(r.Context())

		err := template.CreateEvent(data).Render(r.Context(), w)
		if err != nil {
			panic(err)
		}
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		description := r.FormValue("description")
		from := r.FormValue("from")
		to := r.FormValue("to")

		fromTimestamp, err := time.Parse("2006-01-02T15:04", from)
		if err != nil {
			panic(err)
		}

		toTimestamp, err := time.Parse("2006-01-02T15:04", to)
		if err != nil {
			panic(err)
		}

		newEvent := models.Event{
			Name:        name,
			Description: description,
			Start:       fromTimestamp,
			End:         toTimestamp,
		}

		_, err = rs.db.NewInsert().Model(&newEvent).Returning("*").Exec(r.Context())
		if err != nil {
			panic(err)
		}

		data := template.CreateEventLocationData{
			EventID:   strconv.FormatInt(newEvent.ID, 10),
			Locations: []models.Location{},
		}

		err = rs.db.NewSelect().Model(&data.Locations).Scan(r.Context())
		if err != nil {
			panic(err)
		}

		err = template.CreateEventLocation(data).Render(r.Context(), w)
		if err != nil {
			panic(err)
		}
	})

	r.Post("/{eventID}/location", func(w http.ResponseWriter, r *http.Request) {
		eventID := chi.URLParam(r, "eventID")
		locationID := r.FormValue("location")

		var event models.Event
		err := rs.db.NewSelect().Model(&event).Where("id = ?", eventID).Scan(r.Context())
		if err != nil {
			panic(err)
		}

		// TODO: Add creation of new location

		locationIDInt, err := strconv.ParseInt(locationID, 10, 64)
		if err != nil {
			panic(err)
		}

		event.LocationID = locationIDInt
		_, err = rs.db.NewUpdate().Model(&event).Where("id = ?", eventID).Exec(r.Context())
		if err != nil {
			panic(err)
		}

		data := template.CreateEventCategoryData{
			EventID:    eventID,
			Categories: []models.Category{},
		}

		err = rs.db.NewSelect().Model(&data.Categories).Scan(r.Context())
		if err != nil {
			panic(err)
		}

		err = template.CreateEventCategory(data).Render(r.Context(), w)
	})

	r.Post("/{eventID}/category", func(w http.ResponseWriter, r *http.Request) {
		eventID := chi.URLParam(r, "eventID")
		r.ParseForm()
		categoryIds := r.Form["category"]

		eventIDInt, err := strconv.ParseInt(eventID, 10, 64)
		if err != nil {
			panic(err)
		}

		// Create CategoryToEvent for each category
		for _, categoryID := range categoryIds {
			categoryIDInt, err := strconv.ParseInt(categoryID, 10, 64)
			if err != nil {
				panic(err)
			}

			categoryToEvent := models.CategoryToEvent{
				CategoryID: categoryIDInt,
				EventID:    eventIDInt,
			}

			_, err = rs.db.NewInsert().Model(&categoryToEvent).Exec(r.Context())
			if err != nil {
				panic(err)
			}
		}

		data := template.CreateEventFeesData{
			EventID:      eventID,
			EntranceFees: []models.EntranceFee{},
		}

		err = template.CreateEventFees(data).Render(r.Context(), w)
		if err != nil {
			panic(err)
		}
	})

	r.Post("/{eventID}/fees", func(w http.ResponseWriter, r *http.Request) {
		eventID := chi.URLParam(r, "eventID")
		ticketName := r.FormValue("name")
		ticketPrice := r.FormValue("price")

		eventIDInt, err := strconv.ParseInt(eventID, 10, 64)
		if err != nil {
			panic(err)
		}

		ticketPriceInt, err := strconv.ParseInt(ticketPrice, 10, 64)
		if err != nil {
			panic(err)
		}

		entranceFee := models.EntranceFee{
			Name:    ticketName,
			Price:   ticketPriceInt,
			EventID: eventIDInt,
		}

		_, err = rs.db.NewInsert().Model(&entranceFee).Exec(r.Context())
		if err != nil {
			panic(err)
		}

		var entranceFees []models.EntranceFee
		err = rs.db.NewSelect().Model(&entranceFees).Where("event_id = ?", eventID).Scan(r.Context())
		if err != nil {
			panic(err)
		}

		data := template.CreateEventFeesData{
			EventID:      eventID,
			EntranceFees: entranceFees,
		}

		err = template.EventFeesForm(data).Render(r.Context(), w)
		if err != nil {
			panic(err)
		}
	})

	r.Post("/finish", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("HX-Redirect", "/events")
	})

	return r
}
