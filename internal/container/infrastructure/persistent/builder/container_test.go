package builder

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestContainerBuilderChainMethods(t *testing.T) {
	b := NewContainer().
		IDs(1, 2).
		UserIDs(9).
		Names("demo").
		Images("busybox:latest").
		RuntimeIDs("rid-1").
		Statuses("running").
		Order("id DESC").
		ForUpdate()

	if len(b.ID) != 2 || b.ID[0] != 1 || b.ID[1] != 2 {
		t.Fatalf("unexpected IDs: %+v", b.ID)
	}
	if len(b.UserID) != 1 || b.UserID[0] != 9 {
		t.Fatalf("unexpected UserIDs: %+v", b.UserID)
	}
	if len(b.Name) != 1 || b.Name[0] != "demo" {
		t.Fatalf("unexpected Names: %+v", b.Name)
	}
	if len(b.Image) != 1 || b.Image[0] != "busybox:latest" {
		t.Fatalf("unexpected Images: %+v", b.Image)
	}
	if len(b.RuntimeID) != 1 || b.RuntimeID[0] != "rid-1" {
		t.Fatalf("unexpected RuntimeIDs: %+v", b.RuntimeID)
	}
	if len(b.Status) != 1 || b.Status[0] != "running" {
		t.Fatalf("unexpected Statuses: %+v", b.Status)
	}
	if b.OrderBy != "id DESC" || !b.ForUpdateLock {
		t.Fatalf("unexpected order or lock: order=%s lock=%v", b.OrderBy, b.ForUpdateLock)
	}
}

func TestContainerBuilderFormatArg(t *testing.T) {
	b := NewContainer().IDs(9).UserIDs(1).Statuses("created")
	got, err := b.FormatArg()
	if err != nil {
		t.Fatalf("FormatArg returned error: %v", err)
	}
	if !strings.Contains(got, "\"ID\":[9]") {
		t.Fatalf("FormatArg missing ID payload: %s", got)
	}
	if !strings.Contains(got, "\"UserID\":[1]") {
		t.Fatalf("FormatArg missing user id payload: %s", got)
	}
	if !strings.Contains(got, "\"Status\":[\"created\"]") {
		t.Fatalf("FormatArg missing status payload: %s", got)
	}
}

func TestContainerBuilderFillQuery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:container_builder_fill?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	query := NewContainer().
		IDs(3, 4).
		UserIDs(1).
		Names("demo").
		Images("busybox:latest").
		RuntimeIDs("rid").
		Statuses("running").
		Order("id DESC").
		ForUpdate()

	stmt := query.Fill(db.Session(&gorm.Session{DryRun: true})).
		Table("containers").
		Find(&[]map[string]any{}).
		Statement

	sql := stmt.SQL.String()
	if !strings.Contains(sql, "id in") {
		t.Fatalf("expected id filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "user_id in") {
		t.Fatalf("expected user_id filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "name IN") {
		t.Fatalf("expected name filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "image IN") {
		t.Fatalf("expected image filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "runtime_id IN") {
		t.Fatalf("expected runtime_id filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "status IN") {
		t.Fatalf("expected status filter in SQL, got: %s", sql)
	}
	if !strings.Contains(strings.ToLower(sql), "order by") {
		t.Fatalf("expected order by in SQL, got: %s", sql)
	}
}
