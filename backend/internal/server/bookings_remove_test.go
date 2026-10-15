package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// pathBooking собирает путь к записи по её идентификатору.
func pathBooking(id int32) string {
	return fmt.Sprintf("/api/bookings/%d", id)
}

// deleteBooking отправляет DELETE /api/bookings/{id} и возвращает код ответа.
func deleteBooking(t *testing.T, router http.Handler, id int32) int {
	t.Helper()

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, pathBooking(id), nil))

	return rec.Code
}

// TestBookingsRemoveFreesSlot проверяет, что отмена убирает запись из
// предстоящих встреч и освобождает её слот для повторной записи.
func TestBookingsRemoveFreesSlot(t *testing.T) {
	router, db := newSeededRouter(t)
	id := insertEventType(t, db, 15)
	startsAt := tomorrowAt(9, 0)

	code, raw := postBooking(t, router, bookingRequest{
		EventTypeID: int64(id), StartsAt: startsAt, GuestName: "Анна", GuestEmail: "anna@example.com",
	})
    // TODO: check it
	if code != http.StatusCreated {
		t.Fatalf("подготовить запись: ожидался статус %d, получен %d, тело: %s", http.StatusCreated, code, raw)
	}

	bookings := getUpcoming(t, router)
	if len(bookings) != 1 {
		t.Fatalf("до отмены ожидалась 1 встреча, получено %d", len(bookings))
	}
	bookingID := int32(bookings[0].Id)

	if code := deleteBooking(t, router, bookingID); code != http.StatusNoContent {
		t.Fatalf("ожидался статус %d, получен %d", http.StatusNoContent, code)
	}

	if bookings := getUpcoming(t, router); len(bookings) != 0 {
		t.Fatalf("после отмены ожидалось 0 встреч, получено %d", len(bookings))
	}

	_, slots, _ := getSlots(t, router, id, dateString(startsAt))
	free := false
	for _, clock := range slotClock(t, slots.Slots) {
		if clock == "09:00" {
			free = true
		}
	}
	if !free {
		t.Error("слот 09:00 не освободился после отмены записи")
	}
}

// TestBookingsRemoveNotFound проверяет 404 для отсутствующей записи и
// повторной отмены уже отменённой.
func TestBookingsRemoveNotFound(t *testing.T) {
	router, db := newSeededRouter(t)
	id := insertEventType(t, db, 15)

	t.Run("отсутствующая запись", func(t *testing.T) {
		if code := deleteBooking(t, router, 9999); code != http.StatusNotFound {
			t.Fatalf("ожидался статус %d, получен %d", http.StatusNotFound, code)
		}
	})

	t.Run("повторная отмена", func(t *testing.T) {
		code, raw := postBooking(t, router, bookingRequest{
			EventTypeID: int64(id), StartsAt: tomorrowAt(9, 0), GuestName: "Анна", GuestEmail: "anna@example.com",
		})
		if code != http.StatusCreated {
			t.Fatalf("подготовить запись: ожидался статус %d, получен %d, тело: %s", http.StatusCreated, code, raw)
		}

		bookings := getUpcoming(t, router)
		if len(bookings) != 1 {
			t.Fatalf("ожидалась 1 встреча, получено %d", len(bookings))
		}
		bookingID := int32(bookings[0].Id)

		if code := deleteBooking(t, router, bookingID); code != http.StatusNoContent {
			t.Fatalf("первая отмена: ожидался статус %d, получен %d", http.StatusNoContent, code)
		}
		if code := deleteBooking(t, router, bookingID); code != http.StatusNotFound {
			t.Fatalf("повторная отмена: ожидался статус %d, получен %d", http.StatusNotFound, code)
		}
	})
}
