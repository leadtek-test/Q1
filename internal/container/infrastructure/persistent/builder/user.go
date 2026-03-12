package builder

import (
	"github.com/leadtek-test/q1/common/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID       []uint
	Username []string
	Password string

	// extend fields
	OrderBy       string
	ForUpdateLock bool
}

func NewUser() *User {
	return &User{}
}

func (u *User) FormatArg() (string, error) {
	return util.MarshalString(u)
}

func (u *User) Fill(db *gorm.DB) *gorm.DB {
	db = u.fillWhere(db)
	if u.OrderBy != "" {
		db = db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: u.OrderBy},
		})
	}
	return db
}

func (u *User) fillWhere(db *gorm.DB) *gorm.DB {
	if len(u.ID) > 0 {
		db = db.Where("id in (?)", u.IDs)
	}
	if len(u.Username) > 0 {
		db = db.Where("username IN (?)", u.Username)
	}
	if u.Password != "" {
		db = db.Where("password = ?", u.Password)
	}
	if u.ForUpdateLock {
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}

func (u *User) IDs(v ...uint) *User {
	u.ID = v
	return u
}

func (u *User) Usernames(v ...string) *User {
	u.Username = v
	return u
}

func (u *User) Passwords(v string) *User {
	u.Password = v
	return u
}

func (u *User) Order(v string) *User {
	u.OrderBy = v
	return u
}

func (u *User) ForUpdate() *User {
	u.ForUpdateLock = true
	return u
}
