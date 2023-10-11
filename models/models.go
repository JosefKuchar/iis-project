package models

import (
	"time"

	"github.com/uptrace/bun"
)

type UserToEvent struct {
	bun.BaseModel `bun:"table:user_to_event"`

	UserID  int64  `bun:",pk"`
	User    *User  `bun:"rel:belongs-to"`
	EventID int64  `bun:",pk"`
	Event   *Event `bun:"rel:belongs-to"`
}

type CategoryToEvent struct {
	bun.BaseModel `bun:"table:category_to_event"`

	CategoryID int64     `bun:",pk"`
	Category   *Category `bun:"rel:belongs-to"`
	EventID    int64     `bun:",pk"`
	Event      *Event    `bun:"rel:belongs-to"`
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID       int64  `bun:",pk,autoincrement"`
	Name     string `faker:"name"`
	Email    string
	Password string
	RoleID   int64
	Role     *Role   `bun:"rel:belongs-to"`
	Events   []Event `bun:"m2m:user_to_event,join:User=Event"`
}

type Role struct {
	bun.BaseModel `bun:"table:roles"`

	ID   int64  `bun:",pk,autoincrement"`
	Name string `bun:",notnull"`
}

type Category struct {
	bun.BaseModel `bun:"table:categories"`

	ID         int64      `bun:",pk,autoincrement"`
	Name       string     `bun:",notnull"`
	ParentID   int64      `bun:",nullzero"`
	Parent     *Category  `bun:"rel:belongs-to"`
	Categories []Category `bun:"rel:has-many"`
	Approved   bool       `bun:",notnull"`
}

type Location struct {
	bun.BaseModel `bun:"table:locations"`

	ID       int64  `bun:",pk,autoincrement"`
	Name     string `bun:",notnull"`
	Street   string `bun:",notnull"`
	Zip      string `bun:",notnull"`
	City     string `bun:",notnull"`
	Approved bool   `bun:",notnull"`
}

type Event struct {
	bun.BaseModel `bun:"table:events"`

	ID           int64  `bun:",pk,autoincrement"`
	Name         string `bun:",notnull"`
	LocationID   int64
	Location     *Location     `bun:"rel:belongs-to"`
	Description  string        `bun:","`
	Start        time.Time     `bun:",notnull"`
	End          time.Time     `bun:",notnull"`
	Capacity     int64         `bun:","`
	EntranceFees []EntranceFee `bun:"rel:has-many"`
	Comments     []Comment     `bun:"rel:has-many"`
	Ratings      []Rating      `bun:"rel:has-many"`
	Categories   []Category    `bun:"m2m:category_to_event,join:Event=Category"`
	Approved     bool          `bun:",notnull"`
}

type EntranceFee struct {
	bun.BaseModel `bun:"table:entrance_fees"`

	ID      int64 `bun:",pk,autoincrement"`
	EventID int64
	Event   *Event `bun:"rel:belongs-to"`
	Price   int64  `bun:",notnull"`
	Name    string `bun:",notnull"`
}

type Comment struct {
	bun.BaseModel `bun:"table:comments"`

	ID      int64 `bun:",pk,autoincrement"`
	EventID int64
	Event   *Event `bun:"rel:belongs-to"`
	UserID  int64
	User    *User  `bun:"rel:belongs-to"`
	Text    string `bun:",notnull"`
}

type Rating struct {
	bun.BaseModel `bun:"table:ratings"`

	ID      int64 `bun:",pk,autoincrement"`
	EventID int64
	Event   *Event `bun:"rel:belongs-to"`
	UserID  int64
	User    *User  `bun:"rel:belongs-to"`
	Text    string `bun:",notnull"`
	Rating  int64  `bun:",notnull"`
}
