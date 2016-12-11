package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/igormartire/gorfiv/models"
)

const (
	apiToken = "sweetpotato"
)

type MockRepo struct {
	GetInvoiceById_Called         bool
	GetInvoiceById_ParameterValue int
	GetInvoiceById_ReturnValue    *models.Invoice
	GetInvoiceById_ReturnError    error
}

func (r *MockRepo) GetInvoices(opts *models.QueryOptions) (invoices []*models.Invoice, err error) {
	return nil, nil
}
func (r *MockRepo) GetInvoiceById(id int) (*models.Invoice, error) {
	r.GetInvoiceById_Called = true
	r.GetInvoiceById_ParameterValue = id
	return r.GetInvoiceById_ReturnValue, r.GetInvoiceById_ReturnError
}
func (r *MockRepo) InsertInvoice(i models.Invoice) (id int64, err error) {
	return 0, nil
}
func (r *MockRepo) DeleteInvoice(id int) (nRows int64, err error) {
	return 0, nil
}
func (r *MockRepo) UpdateInvoice(id int, newDescription string) (nRows int64, err error) {
	return 0, nil
}
func (r *MockRepo) CountInvoices(opts *models.QueryOptions) (count int, err error) {
	return 0, nil
}

var invoiceStub = models.Invoice{
	Id:             1,
	Document:       "docStub",
	Description:    "descriptionStub",
	Amount:         42.42,
	CreatedAt:      time.Now(),
	ReferenceMonth: int(time.Now().Month()),
	ReferenceYear:  time.Now().Year(),
	IsActive:       true,
}

// This function is used for setup before executing the test functions
func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func TestUnauthenticatedWithoutToken(t *testing.T) {
	server := New(&Env{}, apiToken)
	var routes = []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/invoices"},
		{"GET", "/invoices/1"},
		{"POST", "/invoices"},
		{"PUT", "/invoices/1"},
		{"DELETE", "/invoices/1"},
	}

	for _, route := range routes {
		req, err := http.NewRequest(route.method, route.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		assert := newAssert(t, route, w)
		assert.StatusCodeEquals(http.StatusUnauthorized)
		assert.BodyErrorMessageEquals("API token required")
	}
}

func TestUnauthenticatedWithWrongToken(t *testing.T) {
	server := New(&Env{}, apiToken)
	var routes = []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/invoices"},
		{"GET", "/invoices/1"},
		{"POST", "/invoices"},
		{"PUT", "/invoices/1"},
		{"DELETE", "/invoices/1"},
	}

	for _, route := range routes {
		req, err := http.NewRequest(route.method, route.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := req.URL.Query()
		q.Add("apiToken", "wrongApiToken")
		req.URL.RawQuery = q.Encode()
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		assert := newAssert(t, route, w)
		assert.StatusCodeEquals(http.StatusUnauthorized)
		assert.BodyErrorMessageEquals("Invalid API token")
	}
}

func TestAuthenticated(t *testing.T) {
	repo := &MockRepo{}
	server := New(NewEnv(repo), apiToken)
	var routes = []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/invoices"},
		{"GET", "/invoices/1"},
		{"POST", "/invoices"},
		{"PUT", "/invoices/1"},
		{"DELETE", "/invoices/1"},
	}

	for _, route := range routes {
		req, err := http.NewRequest(route.method, route.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := req.URL.Query()
		q.Add("apiToken", apiToken)
		req.URL.RawQuery = q.Encode()
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		if w.Result().StatusCode == http.StatusUnauthorized {
			t.Errorf("Route %v: status code should have been %v, but was %v instead.", route, http.StatusUnauthorized, w.Result().StatusCode)
		}
	}
}

func TestRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("apiToken", apiToken)
	req.URL.RawQuery = q.Encode()
	repo := &MockRepo{}
	server := New(NewEnv(repo), apiToken)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	assert := newAssert(t, "GET /", w)
	assert.StatusCodeEquals(http.StatusMovedPermanently)
}

func TestInvoicesShowNonIntegerId(t *testing.T) {
	req, err := http.NewRequest("GET", "/invoices/nan?apiToken="+apiToken, nil)
	if err != nil {
		t.Fatal(err)
	}
	server := New(&Env{}, apiToken)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	assert := newAssert(t, "GET /invoices/nan", w)
	assert.StatusCodeEquals(http.StatusBadRequest)
	assert.BodyErrorMessageEquals("parameter id should be an integer")
}

func TestInvoicesShowUnexistentId(t *testing.T) {
	req, err := http.NewRequest("GET", "/invoices/1?apiToken="+apiToken, nil)
	if err != nil {
		t.Fatal(err)
	}
	repo := &MockRepo{}
	repo.GetInvoiceById_ReturnValue = nil
	repo.GetInvoiceById_ReturnError = models.InvoiceNotFound
	server := New(NewEnv(repo), apiToken)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	assert := newAssert(t, "GET /invoices/1", w)
	assert.IsTrue(repo.GetInvoiceById_Called)
	assert.IntEquals(repo.GetInvoiceById_ParameterValue, 1)
	assert.StatusCodeEquals(http.StatusNotFound)
	assert.BodyErrorMessageEquals("there is no resource with the specified id")
}

func TestInvoicesShowWithError(t *testing.T) {
	req, err := http.NewRequest("GET", "/invoices/1?apiToken="+apiToken, nil)
	if err != nil {
		t.Fatal(err)
	}
	repo := &MockRepo{}
	repo.GetInvoiceById_ReturnValue = nil
	repo.GetInvoiceById_ReturnError = errors.New("error")
	server := New(NewEnv(repo), apiToken)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	assert := newAssert(t, "GET /invoices/1", w)
	assert.IsTrue(repo.GetInvoiceById_Called)
	assert.IntEquals(repo.GetInvoiceById_ParameterValue, 1)
	assert.StatusCodeEquals(http.StatusInternalServerError)
}

func TestInvoicesShowSuccess(t *testing.T) {
	req, err := http.NewRequest("GET", "/invoices/1?apiToken="+apiToken, nil)
	if err != nil {
		t.Fatal(err)
	}
	repo := &MockRepo{}
	repo.GetInvoiceById_ReturnValue = &invoiceStub
	repo.GetInvoiceById_ReturnError = nil
	server := New(NewEnv(repo), apiToken)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	assert := newAssert(t, "GET /invoices/1", w)
	assert.IsTrue(repo.GetInvoiceById_Called)
	assert.IntEquals(repo.GetInvoiceById_ParameterValue, 1)
	assert.StatusCodeEquals(http.StatusOK)
}

type assert struct {
	t  *testing.T
	id interface{}
	w  *httptest.ResponseRecorder
}

func newAssert(t *testing.T, id interface{}, w *httptest.ResponseRecorder) *assert {
	return &assert{
		t:  t,
		id: id,
		w:  w,
	}
}

func (a *assert) IsTrue(b bool) {
	if !b {
		a.t.Errorf("%v: expected \"true\", but received \"false\" instead.", a.id)
	}
}

func (a *assert) IntEquals(count int, expectedCount int) {
	if count != expectedCount {
		a.t.Errorf("%v: count should have been %v, but was %v instead.", a.id, expectedCount, count)
	}
}

func (a *assert) StatusCodeEquals(expectedStatusCode int) {
	if a.w.Result().StatusCode != expectedStatusCode {
		a.t.Errorf("%v: status code should have been %v, but was %v instead.", a.id, expectedStatusCode, a.w.Result().StatusCode)
	}
}

func (a *assert) BodyErrorMessageEquals(expectedErrorMessage string) {
	if a.w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		a.t.Errorf("%v: wrong Content-Type header. Expected \"application/json; charset=utf-8\", but was \"%v\" instead.", a.id, a.w.Header().Get("Content-Type"))
	}
	var body struct {
		Error string
	}
	err := json.Unmarshal(a.w.Body.Bytes(), &body)
	if err != nil {
		a.t.Error(err)
	} else if body.Error != expectedErrorMessage {
		a.t.Errorf("%v: response's body error should have been \"%v\", but was \"%v\" instead.", a.id, expectedErrorMessage, body.Error)
	}
}

func (a *assert) JsonItemEquals(i *models.Invoice) {
	var response struct {
		Item models.Invoice
	}
	err := json.Unmarshal(a.w.Body.Bytes(), &response)
	if err != nil {
		a.t.Error(err)
	} else if !(&response.Item).Equals(i) {
		a.t.Errorf("%v: returned invoice should have been %v, but was %v instead.", a.id, i, response.Item)
	}
}
