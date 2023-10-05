package routes

import (
	"database/sql"
	"net/http"

	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/settings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type resources struct {
	db *bun.DB
}

func (rs resources) Routes() chi.Router {
	r := chi.NewRouter()
	r.Mount("/login", rs.LoginRoutes())
	r.Mount("/register", rs.RegisterRoutes())
	r.Mount("/logout", rs.LogoutRoutes())
	r.Mount("/events", rs.EventRoutes())
	r.Mount("/user", rs.UserRoutes())

	r.Group(func(r chi.Router) {
		// All logged in users
		r.Group(func(r chi.Router) {
			r.Use(UserAuthenticator)
			r.Mount("/create-event", rs.CreateEventRoutes())
		})
		// Moderators and admins
		r.Group(func(r chi.Router) {
			r.Use(ModeratorAuthenticator)
			r.Mount("/admin/events", rs.AdminEventsRoutes())
			r.Mount("/admin/categories", rs.AdminCategoriesRoutes())
			r.Mount("/admin/comments", rs.AdminCommentsRoutes())
			r.Mount("/admin/locations", rs.AdminLocationsRoutes())
		})
		// Admins only
		r.Group(func(r chi.Router) {
			r.Use(AdminAuthenticator)
			r.Mount("/admin/users", rs.AdminUsersRoutes())
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Redirect to events
		http.Redirect(w, r, "/events", http.StatusMovedPermanently)
	})

	return r
}

var tokenAuth *jwtauth.JWTAuth

func init() {
	// TODO: change secret
	// Generate token
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

func Router() chi.Router {
	// Connect to database
	sqldb, err := sql.Open("mysql", "root:@/iis")
	if err != nil {
		panic(err)
	}
	db := bun.NewDB(sqldb, mysqldialect.New())
	db.RegisterModel(
		(*models.UserToEvent)(nil),
		(*models.CategoryToEvent)(nil),
		(*models.User)(nil),
		(*models.Role)(nil),
		(*models.Category)(nil),
		(*models.Location)(nil),
		(*models.Event)(nil),
		(*models.EntranceFee)(nil),
		(*models.Comment)(nil),
		(*models.Rating)(nil),
	)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))

	// Create router
	r := chi.NewRouter()

	// db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	resources := &resources{db: db}

	// Set up routes
	r.Use(middleware.Logger)
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Handle("/*", resources.Routes())

	return r
}

// https://github.com/go-chi/jwtauth/blob/master/jwtauth.go

func UserAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

func ModeratorAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if user is moderator or admin
		if int(claims["RoleID"].(float64)) != settings.ROLE_MODERATOR && int(claims["RoleID"].(float64)) != settings.ROLE_ADMIN {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

func AdminAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if user is admin
		if int(claims["RoleID"].(float64)) != settings.ROLE_ADMIN {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}
