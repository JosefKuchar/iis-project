package main

import (
	"JosefKuchar/iis-project/models"
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func main() {
	godotenv.Load()
	mysqldn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")
	sqldb, err := sql.Open("mysql", mysqldn)
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

	// Temporary disable foreign keys
	db.NewRaw("SET FOREIGN_KEY_CHECKS = 0").Exec(context.Background())

	ctx := context.Background()
	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	if err := fixture.Load(ctx, os.DirFS("cmd"), "fixtures.yaml"); err != nil {
		panic(err)
	}

	// Enable foreign keys
	db.NewRaw("SET FOREIGN_KEY_CHECKS = 1").Exec(context.Background())

	for i := 0; i < 10; i++ {
		location := &models.Location{
			Name:     gofakeit.Name(),
			Street:   gofakeit.Street(),
			Zip:      gofakeit.Zip(),
			City:     gofakeit.City(),
			Approved: true,
		}
		db.NewInsert().Model(location).Exec(ctx)
	}

	for i := 0; i < 10; i++ {
		start := gofakeit.Date()
		event := &models.Event{
			Name:        gofakeit.Word(),
			Description: gofakeit.LoremIpsumSentence(20),
			Capacity:    int64(gofakeit.Number(1, 50)),
			Start:       start,
			End:         start.Add(time.Hour * time.Duration(gofakeit.Number(1, 10))),
			LocationID:  int64(gofakeit.Number(1, 10)),
			Approved:    true,
		}
		db.NewInsert().Model(event).Exec(ctx)

		for j := 0; j < 3; j++ {
			entranceFee := &models.EntranceFee{
				EventID: int64(i + 1),
				Name:    gofakeit.Name(),
				Price:   int64(gofakeit.Number(1, 100)),
			}
			db.NewInsert().Model(entranceFee).Exec(ctx)
		}

		for j := 0; j < 3; j++ {
			categoryToEvent := &models.CategoryToEvent{
				EventID:    int64(i + 1),
				CategoryID: int64(gofakeit.Number(1, 6)),
			}
			db.NewInsert().Model(categoryToEvent).Exec(ctx)
		}
	}

	for i := 0; i < 100; i++ {
		user := &models.User{
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			Password: "$2a$10$Ioo2eOK1UZSJiQ2oh.4Unuvl7MHtrQZLA8WEnEZDytacLXqDFoAXS", // "password"
			RoleID:   int64(gofakeit.Number(1, 3)),
		}
		db.NewInsert().Model(user).Exec(ctx)

		for j := 0; j < 3; j++ {
			// Fetch random event and fees
			var event models.Event
			db.NewSelect().Model(&event).Relation("EntranceFees").Relation("Categories").Where("id = ?", int64(gofakeit.Number(1, 10))).Scan(ctx)

			// Create userToEvent with random entrance fee
			userToEvent := &models.UserToEvent{
				UserID:        int64(i + 1),
				EventID:       event.ID,
				EntranceFeeID: event.EntranceFees[gofakeit.Number(0, len(event.EntranceFees)-1)].ID,
			}
			db.NewInsert().Model(userToEvent).Exec(ctx)
		}
	}

	for i := 0; i < 10; i++ {
		comment := &models.Comment{
			Text:    gofakeit.LoremIpsumSentence(20),
			EventID: int64(gofakeit.Number(1, 10)),
			UserID:  int64(gofakeit.Number(1, 10)),
		}
		db.NewInsert().Model(comment).Exec(ctx)
	}

	for i := 0; i < 10; i++ {
		rating := &models.Rating{
			Rating:  int64(gofakeit.Number(1, 5)),
			EventID: int64(gofakeit.Number(1, 10)),
			UserID:  int64(gofakeit.Number(1, 10)),
		}
		db.NewInsert().Model(rating).Exec(ctx)
	}

	/* FOREIGN KEYS */
	// UserToEvent
	db.NewRaw("ALTER TABLE `user_to_event` ADD FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `user_to_event` ADD FOREIGN KEY (`event_id`) REFERENCES `events`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `user_to_event` ADD FOREIGN KEY (`entrance_fee_id`) REFERENCES `entrance_fees`(`id`) ON DELETE CASCADE").Exec(ctx)

	// CategoryToEvent
	db.NewRaw("ALTER TABLE `category_to_event` ADD FOREIGN KEY (`category_id`) REFERENCES `categories`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `category_to_event` ADD FOREIGN KEY (`event_id`) REFERENCES `events`(`id`) ON DELETE CASCADE").Exec(ctx)

	// User
	db.NewRaw("ALTER TABLE `users` ADD FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE").Exec(ctx)

	// Category
	db.NewRaw("ALTER TABLE `categories` ADD FOREIGN KEY (`parent_id`) REFERENCES `categories`(`id`) ON DELETE CASCADE").Exec(ctx)

	// Event
	db.NewRaw("ALTER TABLE `events` ADD FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `events` ADD FOREIGN KEY (`owner_id`) REFERENCES `users`(`id`) ON DELETE CASCADE").Exec(ctx)

	// EntranceFee
	db.NewRaw("ALTER TABLE `entrance_fees` ADD FOREIGN KEY (`event_id`) REFERENCES `events`(`id`) ON DELETE CASCADE").Exec(ctx)

	// Comment
	db.NewRaw("ALTER TABLE `comments` ADD FOREIGN KEY (`event_id`) REFERENCES `events`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `comments` ADD FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE").Exec(ctx)

	// Rating
	db.NewRaw("ALTER TABLE `ratings` ADD FOREIGN KEY (`event_id`) REFERENCES `events`(`id`) ON DELETE CASCADE").Exec(ctx)
	db.NewRaw("ALTER TABLE `ratings` ADD FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE").Exec(ctx)
}
