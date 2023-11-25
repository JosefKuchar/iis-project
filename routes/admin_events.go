package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/uptrace/bun"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminEventsRoutes() chi.Router {
	r := chi.NewRouter()

	parseForm := func(r *http.Request) (template.AdminEventPageData, error) {
		data := template.AdminEventPageData{}
		data.Errors = make(map[string]string)

		idString := chi.URLParam(r, "id")
		if idString == "" {
			idString = "0"
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			return data, err
		}

		r.ParseForm()

		data.Event.ID = int64(id)
		data.Event.Name = r.FormValue("name")
		if r.FormValue("capacity") != "" {
			capacity, err := strconv.Atoi(r.FormValue("capacity"))
			if err != nil {
				return data, err
			}
			data.Event.Capacity = int64(capacity)
		}
		if r.FormValue("start") != "" {
			from, err := time.Parse("2006-01-02T15:04", r.FormValue("start"))
			if err != nil {
				return data, err
			}
			data.Event.Start = from
		}
		if r.FormValue("end") != "" {
			from, err := time.Parse("2006-01-02T15:04", r.FormValue("end"))
			if err != nil {
				return data, err
			}
			data.Event.End = from
		}
		data.Event.Description = r.FormValue("description")
		data.Event.Approved = r.FormValue("approved") == "on"
		if r.FormValue("location_id") != "" {
			location, err := strconv.Atoi(r.FormValue("location_id"))
			if err != nil {
				return data, err
			}
			data.Event.LocationID = int64(location)
		}
		data.New = r.FormValue("new") == "true"

		// Entrance fees
		entranceFeeNames := r.Form["entranceFeeName[]"]
		entranceFeePrices := r.Form["entranceFeePrice[]"]
		entranceFeeIDs := r.Form["entranceFeeID[]"]
		for i := range entranceFeeNames {
			id, err := strconv.Atoi(entranceFeeIDs[i])
			if err != nil {
				return data, err
			}
			price := 0
			if entranceFeePrices[i] != "" {
				price, err = strconv.Atoi(entranceFeePrices[i])
				if err != nil {
					return data, err
				}
			}

			data.Event.EntranceFees = append(data.Event.EntranceFees, models.EntranceFee{
				ID:    int64(id),
				Name:  entranceFeeNames[i],
				Price: int64(price),
			})
		}

		// Categories
		categoryIDs := r.Form["categories[]"]
		for i := range categoryIDs {
			id, err := strconv.Atoi(categoryIDs[i])
			if err != nil {
				return data, err
			}
			data.Event.Categories = append(data.Event.Categories, models.Category{
				ID: int64(id),
			})
		}

		if data.Event.Name == "" {
			data.Errors["Name"] = "Name cannot be empty"
		}

		// Get user role
		_, claims, _ := jwtauth.FromContext(r.Context())
		data.UserRole = int(claims["RoleID"].(float64))

		return data, nil
	}

	getListData := func(w *http.ResponseWriter, r *http.Request) (template.AdminEventsPageData, error) {
		page, offset, err := getPageOffset(r)
		if err != nil {
			return template.AdminEventsPageData{}, err
		}

		query := r.FormValue("query")

		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			return template.AdminEventsPageData{}, err
		}

		data := template.AdminEventsPageData{}
		eventsQuery := rs.db.
			NewSelect().
			Model(&data.Events).
			WhereGroup("AND", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.
					Where("event.name like ?", "%"+query+"%").
					WhereOr("event.id like ?", "%"+query+"%")
			}).
			Order("approved ASC").
			Limit(settings.PAGE_SIZE).
			Offset(offset)

		data.UserRole = int(claims["RoleID"].(float64))

		if data.UserRole != settings.ROLE_ADMIN && data.UserRole != settings.ROLE_MODERATOR {
			userId, err := strconv.ParseInt(fmt.Sprintf("%v", claims["ID"]), 10, 64)
			if err != nil {
				return template.AdminEventsPageData{}, err
			}

			eventsQuery = eventsQuery.Where("owner_id = ?", userId)
		}

		count, err := eventsQuery.ScanAndCount(r.Context())

		if err != nil {
			return template.AdminEventsPageData{}, err
		}

		data.TotalCount = count
		data.Page = page
		data.Query = query

		(*w).Header().Set("HX-Push-Url", "/admin/events?page="+strconv.Itoa(page)+"&query="+query)

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

		err = template.AdminEventsPage(data, appbar).Render(r.Context(), w)
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
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminEventsPageTable(data, appbar, true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Post("/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		approved := r.FormValue("approved") == "on"

		_, err = rs.db.
			NewUpdate().
			Model(&models.Event{ID: int64(id), Approved: approved}).
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

		err = template.AdminEventsPageTable(data, appbar, true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// New event detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminEventPageData{}
		data.New = true
		data.Event.Approved = true
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminEventPage(data, appbar).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	syncFees := func(r *http.Request, fees []models.EntranceFee, eventId int) error {
		// Get current entrance fees
		var currentEntranceFees []models.EntranceFee
		err := rs.db.
			NewSelect().
			Model(&currentEntranceFees).
			Where("event_id = ?", eventId).
			Scan(r.Context())
		if err != nil {
			return err
		}
		// Sync entrance fees
		for _, fee := range fees {
			fee.EventID = int64(eventId)
			// If fee is new, insert it
			if fee.ID == 0 {
				_, err := rs.db.NewInsert().Model(&fee).Exec(r.Context())
				if err != nil {
					return err
				}
			} else {
				// Update existing fee
				_, err := rs.db.NewUpdate().Model(&fee).Where("id = ?", fee.ID).Exec(r.Context())
				if err != nil {
					return err
				}
			}
		}
		// Delete removed fees
		for _, fee := range currentEntranceFees {
			found := false
			for _, newFee := range fees {
				if fee.ID == newFee.ID {
					found = true
					break
				}
			}
			if !found {
				_, err = rs.db.NewDelete().Model(&fee).Where("id = ?", fee.ID).Exec(r.Context())
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	syncCategories := func(r *http.Request, categories []models.Category, eventId int) error {
		// Remove all categories
		_, err := rs.db.NewDelete().Model((*models.CategoryToEvent)(nil)).Where("event_id = ?", eventId).Exec(r.Context())
		if err != nil {
			return err
		}
		// Insert new categories
		for _, category := range categories {
			_, err := rs.db.NewInsert().Model(&models.CategoryToEvent{
				EventID:    int64(eventId),
				CategoryID: category.ID,
			}).Exec(r.Context())
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Create new event
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// TODO: Check errors
		// Create new event and get its ID
		_, err = rs.db.NewInsert().Model(&data.Event).Returning("id").Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Sync categories
		err = syncCategories(r, data.Event.Categories, int(data.Event.ID))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Sync entrance fees
		err = syncFees(r, data.Event.EntranceFees, int(data.Event.ID))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/events")
	})

	r.Post("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = rs.db.NewUpdate().
			Model(&data.Event).
			Where("id = ?", data.Event.ID).
			OmitZero().
			Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Sync categories
		err = syncCategories(r, data.Event.Categories, int(data.Event.ID))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Sync entrance fees
		err = syncFees(r, data.Event.EntranceFees, int(data.Event.ID))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/events")
	})

	// Existing event detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminEventPageData{}
		data.New = false

		err := rs.db.
			NewSelect().
			Model(&data.Event).
			Relation("EntranceFees").
			Relation("Categories").
			Relation("Location").
			Relation("UsersToEvent").
			Relation("UsersToEvent.User").
			Where("event.id = ?", chi.URLParam(r, "id")).
			Scan(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Get user role
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data.UserRole = int(claims["RoleID"].(float64))

		err = template.AdminEventPage(data, appbar).Render(r.Context(), w)
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

		_, err = rs.db.NewDelete().Model(&models.Event{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
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

		err = template.AdminEventsPageTable(data, appbar, true).Render(r.Context(), w)
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

		_, err = rs.db.NewDelete().Model(&models.Event{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/events")
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminEventPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Delete entrance fee
	r.Post("/{id}/form/delete-fee/{feeIndex}", func(w http.ResponseWriter, r *http.Request) {
		feeIndex, err := strconv.Atoi(chi.URLParam(r, "feeIndex"))
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Remove entrance fee
		for i := range data.Event.EntranceFees {
			if i == feeIndex {
				data.Event.EntranceFees = append(data.Event.EntranceFees[:i], data.Event.EntranceFees[i+1:]...)
				break
			}
		}

		err = template.AdminEventPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Add entrance fee
	r.Post("/{id}/form/add-fee", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data.Event.EntranceFees = append(data.Event.EntranceFees, models.EntranceFee{})
		err = template.AdminEventPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
