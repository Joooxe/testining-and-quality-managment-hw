package apitests_test

import (
	"net/http"
	"testing"

	"github.com/NZRV-88/TestProject/api-tests/internal/config"
	"github.com/NZRV-88/TestProject/api-tests/internal/contract"
	"github.com/NZRV-88/TestProject/api-tests/internal/orders"
	"github.com/NZRV-88/TestProject/api-tests/internal/testdata"
	allure "github.com/allure-framework/allure-go/commons/gotest"
)

func TestOrders(t *testing.T) {
	baseURL, err := config.BaseURL()
	if err != nil {
		t.Fatal(err)
	}
	schemas, err := contract.Load()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		id    string
		title string
		run   func(*scenario)
	}{
		{"1", "Создать заказ", testCreateOrder},
		{"2", "Отклонить нулевое количество", testRejectZeroQuantity},
		{"3", "Получить карточку заказа", testGetOrder},
		{"4", "Найти заказ в списке", testListOrders},
		{"5", "Получить 404 для неизвестного заказа", testGetUnknownOrder},
		{"6", "Перевести заказ из CREATED в PAID", testPayOrder},
		{"7", "Запретить переход из CREATED в DELIVERED", testRejectStatusTransition},
		{"8", "Отменить заказ и сохранить причину", testCancelOrder},
		{"9", "Повторить отмену с той же причиной", testRepeatCancellation},
		{"10", "Отклонить отмену без причины", testRejectCancellationWithoutReason},
		{"11", "Полностью удалить заказ", testRemoveOrder},
		{"12", "Получить 404 при повторном удалении", testRepeatRemoval},
		{"13", "Отклонить синтаксически некорректный JSON", testMalformedJSON},
	}
	for _, tc := range cases {
		allure.Test(t, tc.id, func(a *allure.Context) {
			s := newScenario(a, baseURL, schemas)
			defer s.cleanup()
			tc.run(s)
		}, allure.WithDisplayName(tc.id+" "+tc.title), allure.WithTestCaseID(tc.id),
			allure.WithSuite("Pizza Delivery API"))
	}
}

func testCreateOrder(s *scenario) {
	input := testdata.NewOrder()
	order, header := s.create(input)

	expectEqual(s.a, "Location", header.Get("Location"), "/orders/"+order.ID)
	expectSummary(s.a, order.Summary, "CREATED")
	expectOrderData(s.a, order, input)
	expectNoCancellation(s.a, order)
}

func testRejectZeroQuantity(s *scenario) {
	input := testdata.NewOrder()
	input.Items[0].Qty = 0

	response := s.postOrder(input)
	s.expectError(response, http.StatusBadRequest)
}

func testGetOrder(s *scenario) {
	input := testdata.NewOrder()
	input.Address.Apartment = nil
	created, _ := s.create(input)

	order := s.get(created.ID, "CREATED")
	expectSummary(s.a, order.Summary, "CREATED")
	expectOrderData(s.a, order, input)
	expectNoCancellation(s.a, order)
}

func testListOrders(s *scenario) {
	id := s.createFixture()

	var matches []orders.Summary
	for _, row := range s.list() {
		if row.ID == id {
			matches = append(matches, row)
		}
	}
	expectEqual(s.a, "number of list entries for "+id, len(matches), 1)
	expectSummary(s.a, matches[0], "CREATED")
}

func testGetUnknownOrder(s *scenario) {
	response := s.receive(s.client.Get(s.a.Context(), testdata.NewUUID()))
	s.expectError(response, http.StatusNotFound)
}

func testPayOrder(s *scenario) {
	id := s.createFixture()

	response := s.receive(s.client.ChangeStatus(s.a.Context(), id, "PAID"))
	updated := readJSON[orders.Order](s, response, http.StatusOK, contract.Order)
	expectEqual(s.a, "order ID", updated.ID, id)
	expectEqual(s.a, "updated status", updated.Status, "PAID")
	s.get(id, "PAID")
}

func testRejectStatusTransition(s *scenario) {
	id := s.createFixture()

	response := s.receive(s.client.ChangeStatus(s.a.Context(), id, "DELIVERED"))
	s.expectError(response, http.StatusConflict)
	s.get(id, "CREATED")
}

func testCancelOrder(s *scenario) {
	id := s.createFixture()

	s.cancel(id, "  "+testdata.CancellationNote+"  ")
	expectCancellation(s.a, s.get(id, "CANCELLED"))
}

func testRepeatCancellation(s *scenario) {
	id := s.createFixture()
	s.cancel(id, testdata.CancellationNote)

	s.cancel(id, "  "+testdata.CancellationNote+"  ")
	expectCancellation(s.a, s.get(id, "CANCELLED"))
}

func testRejectCancellationWithoutReason(s *scenario) {
	id := s.createFixture()

	response := s.receive(s.client.Cancel(s.a.Context(), id, nil))
	s.expectError(response, http.StatusBadRequest)
	expectNoCancellation(s.a, s.get(id, "CREATED"))
}

func testRemoveOrder(s *scenario) {
	id := s.createFixture()

	s.remove(id)
	s.expectError(s.receive(s.client.Get(s.a.Context(), id)), http.StatusNotFound)
	for _, row := range s.list() {
		if row.ID == id {
			s.a.Fatalf("removed order %s still appears in GET /orders", id)
		}
	}
}

func testRepeatRemoval(s *scenario) {
	id := s.createFixture()
	s.remove(id)

	response := s.receive(s.client.Remove(s.a.Context(), id, testdata.CancellationNote))
	s.expectError(response, http.StatusNotFound)
}

func testMalformedJSON(s *scenario) {
	response := s.receive(s.client.CreateRawJSON(s.a.Context(), []byte(`{"address":`)))
	s.expectError(response, http.StatusBadRequest)
}
