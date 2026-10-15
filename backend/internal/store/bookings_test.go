package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

// containsStart сообщает, есть ли среди слотов старт с заданным временем суток.
func containsStart(slots []time.Time, hour, minute int) bool {
	for _, slot := range slots {
		if slot.Hour() == hour && slot.Minute() == minute {
			return true
		}
	}

	return false
}

// TestDeleteBookingFreesSlot проверяет, что жёсткое удаление записи освобождает
// её слот и что повторное удаление возвращает ErrNotFound.
func TestDeleteBookingFreesSlot(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	et, err := CreateEventType(ctx, db, CreateEventTypeParams{Name: "Звонок", DurationMinutes: 15})
	if err != nil {
		t.Fatalf("CreateEventType: %v", err)
	}

	start := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	booking, err := CreateBooking(ctx, db, CreateBookingParams{
		EventTypeID: et.ID,
		StartsAt:    start,
		GuestName:   "Анна",
		GuestEmail:  "anna@example.com",
	}, weekdayNow)
	if err != nil {
		t.Fatalf("CreateBooking: %v", err)
	}

	slots, err := Slots(ctx, db, et.ID, start, weekdayNow)
	if err != nil {
		t.Fatalf("Slots до отмены: %v", err)
	}
	if containsStart(slots, 9, 0) {
		t.Fatal("занятый слот 09:00 остался свободным")
	}

	if err := DeleteBooking(ctx, db, booking.ID); err != nil {
		t.Fatalf("DeleteBooking: %v", err)
	}

	slots, err = Slots(ctx, db, et.ID, start, weekdayNow)
	if err != nil {
		t.Fatalf("Slots после отмены: %v", err)
	}
	if !containsStart(slots, 9, 0) {
		t.Error("слот 09:00 не освободился после отмены записи")
	}

	if err := DeleteBooking(ctx, db, booking.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("повторное удаление: ожидался ErrNotFound, получено %v", err)
	}
}
