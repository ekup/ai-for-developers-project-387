package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// newTestDB открывает временную БД со схемой для тестов хранилища.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	return db
}

// weekdayNow — середина недели (среда), чтобы окно записи и выходные в тестах
// были предсказуемы.
var weekdayNow = time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)

// TestSlotsSkipsWeekends проверяет, что на субботу и воскресенье слотов нет,
// а на будний день они есть.
func TestSlotsSkipsWeekends(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	et, err := CreateEventType(ctx, db, CreateEventTypeParams{Name: "Звонок", DurationMinutes: 15})
	if err != nil {
		t.Fatalf("CreateEventType: %v", err)
	}

	friday := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	saturday := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	sunday := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)

	for _, day := range []time.Time{saturday, sunday} {
		slots, err := Slots(ctx, db, et.ID, day, weekdayNow)
		if err != nil {
			t.Fatalf("Slots(%s): %v", day.Weekday(), err)
		}
		if len(slots) != 0 {
			t.Fatalf("для %s ожидалось 0 слотов, получено %v", day.Weekday(), slots)
		}
	}

	slots, err := Slots(ctx, db, et.ID, friday, weekdayNow)
	if err != nil {
		t.Fatalf("Slots(пятница): %v", err)
	}
	if len(slots) == 0 {
		t.Fatal("для буднего дня ожидались слоты")
	}
}

// TestCreateBookingRejectsWeekend проверяет, что создать запись на выходной
// нельзя даже в обход сетки слотов.
func TestCreateBookingRejectsWeekend(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	et, err := CreateEventType(ctx, db, CreateEventTypeParams{Name: "Звонок", DurationMinutes: 15})
	if err != nil {
		t.Fatalf("CreateEventType: %v", err)
	}

	saturday := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)

	_, err = CreateBooking(ctx, db, CreateBookingParams{
		EventTypeID: et.ID,
		StartsAt:    saturday,
		GuestName:   "Анна",
		GuestEmail:  "anna@example.com",
	}, weekdayNow)

	if !errors.Is(err, ErrUnprocessable) {
		t.Fatalf("ожидалась ErrUnprocessable, получено %v", err)
	}
}
