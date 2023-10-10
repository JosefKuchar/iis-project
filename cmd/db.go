package main

import (
	"JosefKuchar/iis-project/cmd/models"
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func main() {
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

	ctx := context.Background()
	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	if err := fixture.Load(ctx, os.DirFS("cmd"), "fixtures.yaml"); err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		location := &models.Location{
			Name:        gofakeit.Name(),
			Description: gofakeit.LoremIpsumSentence(20),
			Approved:    true,
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
			userToEvent := &models.UserToEvent{
				EventID: int64(gofakeit.Number(1, 10)),
				UserID:  int64(i + 1),
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
			Text:    gofakeit.LoremIpsumSentence(10),
		}
		db.NewInsert().Model(rating).Exec(ctx)
	}
}
