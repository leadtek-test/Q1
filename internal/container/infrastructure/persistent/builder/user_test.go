package builder

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserBuilderChainMethods(t *testing.T) {
	b := NewUser().
		IDs(1, 2).
		Usernames("alice", "bob").
		Passwords("p").
		Order("id").
		ForUpdate()

	if len(b.ID) != 2 || b.ID[0] != 1 || b.ID[1] != 2 {
		t.Fatalf("unexpected IDs: %+v", b.ID)
	}
	if len(b.Username) != 2 || b.Username[0] != "alice" || b.Username[1] != "bob" {
		t.Fatalf("unexpected usernames: %+v", b.Username)
	}
	if b.Password != "p" {
		t.Fatalf("unexpected password: %s", b.Password)
	}
	if b.OrderBy != "id" || !b.ForUpdateLock {
		t.Fatalf("unexpected order or lock: order=%s lock=%v", b.OrderBy, b.ForUpdateLock)
	}
}

func TestUserBuilderFormatArg(t *testing.T) {
	b := NewUser().IDs(9).Usernames("alice").Passwords("secret")
	got, err := b.FormatArg()
	if err != nil {
		t.Fatalf("FormatArg returned error: %v", err)
	}
	if !strings.Contains(got, "\"ID\":[9]") {
		t.Fatalf("FormatArg missing ID payload: %s", got)
	}
	if !strings.Contains(got, "\"Username\":[\"alice\"]") {
		t.Fatalf("FormatArg missing username payload: %s", got)
	}
}

func TestUserBuilderFillQuery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:user_builder_fill?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	query := NewUser().
		IDs(3, 4).
		Usernames("alice").
		Passwords("hashed").
		Order("id").
		ForUpdate()

	stmt := query.Fill(db.Session(&gorm.Session{DryRun: true})).
		Table("users").
		Find(&[]map[string]any{}).
		Statement

	sql := stmt.SQL.String()
	if !strings.Contains(sql, "id in") {
		t.Fatalf("expected id filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "username IN") {
		t.Fatalf("expected username filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "password =") {
		t.Fatalf("expected password filter in SQL, got: %s", sql)
	}
	if !strings.Contains(strings.ToLower(sql), "order by") {
		t.Fatalf("expected order by in SQL, got: %s", sql)
	}
}
