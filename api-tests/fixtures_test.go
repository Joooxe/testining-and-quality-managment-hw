package apitests_test

import (
	"encoding/json"
	"net/http"

	"github.com/NZRV-88/TestProject/api-tests/internal/config"
	"github.com/NZRV-88/TestProject/api-tests/internal/contract"
	"github.com/NZRV-88/TestProject/api-tests/internal/orders"
	"github.com/NZRV-88/TestProject/api-tests/internal/testdata"
	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/allure-framework/allure-go/commons/httpexchange"
)

type scenario struct {
	a         *allure.Context
	client    *orders.Client
	schemas   contract.Schemas
	createdID string
}

func newScenario(a *allure.Context, baseURL string, schemas contract.Schemas) *scenario {
	client := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: httpexchange.NewTransport(a.Context(), http.DefaultTransport,
			httpexchange.WithBodyLimit(64<<10)),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &scenario{a: a, client: orders.NewClient(baseURL, client), schemas: schemas}
}

func (s *scenario) receive(response orders.Response, err error) orders.Response {
	s.a.T().Helper()
	if err != nil {
		s.a.Fatalf("HTTP request failed: %v", err)
	}
	return response
}

func (s *scenario) postOrder(input orders.NewOrder) orders.Response {
	s.a.T().Helper()
	response := s.receive(s.client.Create(s.a.Context(), input))
	if response.StatusCode != http.StatusCreated {
		return response
	}

	var identity struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body, &identity); err == nil {
		s.createdID = identity.ID
	}
	return response
}

func (s *scenario) create(input orders.NewOrder) (orders.Order, http.Header) {
	s.a.T().Helper()
	response := s.postOrder(input)
	return readJSON[orders.Order](s, response, http.StatusCreated, contract.Order), response.Header
}

func (s *scenario) createFixture() string {
	s.a.T().Helper()
	order, _ := s.create(testdata.NewOrder())
	return order.ID
}

func (s *scenario) get(id, status string) orders.Order {
	s.a.T().Helper()
	response := s.receive(s.client.Get(s.a.Context(), id))
	order := readJSON[orders.Order](s, response, http.StatusOK, contract.Order)
	expectEqual(s.a, "order ID", order.ID, id)
	expectEqual(s.a, "stored status", order.Status, status)
	return order
}

func (s *scenario) list() []orders.Summary {
	s.a.T().Helper()
	response := s.receive(s.client.List(s.a.Context()))
	return readJSON[[]orders.Summary](s, response, http.StatusOK, contract.List)
}

func (s *scenario) cancel(id, reason string) {
	s.a.T().Helper()
	response := s.receive(s.client.Cancel(s.a.Context(), id, &reason))
	expectNoContent(s.a, response)
}

func (s *scenario) remove(id string) {
	s.a.T().Helper()
	response := s.receive(s.client.Remove(s.a.Context(), id, testdata.CancellationNote))
	expectNoContent(s.a, response)
	s.createdID = ""
}

func (s *scenario) cleanup() {
	if s.createdID == "" {
		return
	}
	response, err := s.client.Remove(s.a.Context(), s.createdID, testdata.CancellationNote)
	if err != nil {
		s.a.Errorf("cleanup order %s: %v", s.createdID, err)
		return
	}
	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound {
		s.a.Errorf("cleanup order %s: HTTP %d, body: %s", s.createdID, response.StatusCode, response.Body)
	}
}
