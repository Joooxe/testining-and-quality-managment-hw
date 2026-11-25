package apitests_test

import (
	"encoding/json"
	"math/big"
	"mime"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/NZRV-88/TestProject/api-tests/internal/contract"
	"github.com/NZRV-88/TestProject/api-tests/internal/orders"
	"github.com/NZRV-88/TestProject/api-tests/internal/testdata"
	allure "github.com/allure-framework/allure-go/commons/gotest"
)

func expectEqual[T comparable](a *allure.Context, field string, got, want T) {
	a.T().Helper()
	if got != want {
		a.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

func expectStatus(a *allure.Context, response orders.Response, want int) {
	a.T().Helper()
	if response.StatusCode != want {
		a.Fatalf("%s: HTTP %d, want %d, body: %s", response.Path, response.StatusCode, want, response.Body)
	}
}

func expectNoContent(a *allure.Context, response orders.Response) {
	a.T().Helper()
	expectStatus(a, response, http.StatusNoContent)
	expectEqual(a, "response body length", len(response.Body), 0)
}

func readJSON[T any](s *scenario, response orders.Response, status int, shape contract.Shape) T {
	s.a.T().Helper()
	expectStatus(s.a, response, status)

	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		s.a.Fatalf("Content-Type: got %q, want application/json", contentType)
	}
	if err := s.schemas.Validate(shape, response.Body); err != nil {
		s.a.Fatalf("response does not match the %s schema: %v", shape, err)
	}

	var value T
	if err := json.Unmarshal(response.Body, &value); err != nil {
		s.a.Fatalf("decode %s response: %v", shape, err)
	}
	return value
}

func (s *scenario) expectError(response orders.Response, status int) {
	s.a.T().Helper()
	body := readJSON[orders.APIError](s, response, status, contract.Error)
	expectEqual(s.a, "JSON error status", body.Status, status)
	expectEqual(s.a, "JSON error path", body.Path, response.Path)
}

func expectSummary(a *allure.Context, summary orders.Summary, status string) {
	a.T().Helper()
	expectEqual(a, "status", summary.Status, status)
	expectEqual(a, "currency", summary.Currency, "EUR")
	expectEqual(a, "deliveryZone", summary.DeliveryZone, "CENTER")

	total, ok := new(big.Rat).SetString(summary.TotalAmount.String())
	if !ok || total.Cmp(big.NewRat(testdata.ExpectedTotal, 1)) != 0 {
		a.Fatalf("totalAmount: got %s, want %d", summary.TotalAmount, testdata.ExpectedTotal)
	}
}

func expectOrderData(a *allure.Context, got orders.Order, want orders.NewOrder) {
	a.T().Helper()
	if !reflect.DeepEqual(got.Address, want.Address) {
		a.Fatalf("address: got %+v, want %+v", got.Address, want.Address)
	}
	if !slices.Equal(sortedItems(got.Items), sortedItems(want.Items)) {
		a.Fatalf("items: got %+v, want %+v", got.Items, want.Items)
	}
}

func sortedItems(items []orders.Item) []orders.Item {
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b orders.Item) int {
		return strings.Compare(a.Name, b.Name)
	})
	return sorted
}

func expectCancellation(a *allure.Context, order orders.Order) {
	a.T().Helper()
	if order.CancellationReason == nil {
		a.Fatal("cancellationReason is null")
	}
	expectEqual(a, "cancellationReason", *order.CancellationReason, testdata.CancellationNote)
}

func expectNoCancellation(a *allure.Context, order orders.Order) {
	a.T().Helper()
	if order.CancellationReason != nil {
		a.Fatalf("cancellationReason: got %q, want null", *order.CancellationReason)
	}
}
