package builder

import (
	"github.com/leadtek-test/q1/common/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Container struct {
	ID        []uint
	UserID    []uint
	Name      []string
	Image     []string
	RuntimeID []string
	Status    []string

	// extend fields
	OrderBy       string
	ForUpdateLock bool
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) FormatArg() (string, error) {
	return util.MarshalString(c)
}

func (c *Container) Fill(db *gorm.DB) *gorm.DB {
	db = c.fillWhere(db)
	if c.OrderBy != "" {
		db = db.Order(c.OrderBy)
	}
	return db
}

func (c *Container) fillWhere(db *gorm.DB) *gorm.DB {
	if len(c.ID) > 0 {
		db = db.Where("id in (?)", c.ID)
	}
	if len(c.UserID) > 0 {
		db = db.Where("user_id in (?)", c.UserID)
	}
	if len(c.Name) > 0 {
		db = db.Where("name IN (?)", c.Name)
	}
	if len(c.Image) > 0 {
		db = db.Where("image IN (?)", c.Image)
	}
	if len(c.RuntimeID) > 0 {
		db = db.Where("runtime_id IN (?)", c.RuntimeID)
	}
	if len(c.Status) > 0 {
		db = db.Where("status IN (?)", c.Status)
	}
	if c.ForUpdateLock {
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}

func (c *Container) IDs(v ...uint) *Container {
	c.ID = v
	return c
}

func (c *Container) UserIDs(v ...uint) *Container {
	c.UserID = v
	return c
}

func (c *Container) Names(v ...string) *Container {
	c.Name = v
	return c
}

func (c *Container) Images(v ...string) *Container {
	c.Image = v
	return c
}

func (c *Container) RuntimeIDs(v ...string) *Container {
	c.RuntimeID = v
	return c
}

func (c *Container) Statuses(v ...string) *Container {
	c.Status = v
	return c
}

func (c *Container) Order(v string) *Container {
	c.OrderBy = v
	return c
}

func (c *Container) ForUpdate() *Container {
	c.ForUpdateLock = true
	return c
}
