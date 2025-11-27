package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleCalculate(t *testing.T) {
	reqBody := CalculateRequest{
		Name:        "Test User",
		BirthDate:   "1980-01-01",
		HireDate:    "2005-01-01",
		Salary:      100000,
		High3:       95000,
		TSPTrad:     500000,
		TSPRoth:     50000,
		TSPContrib:  0.05,
		RetireDate:  "2040-01-01",
		State:       "Pennsylvania",
		SSAge:       67,
		SSEstimate:  2500,
		TSPStrategy: "4_percent_rule",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleCalculate(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp CalculateResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)

	assert.Equal(t, "2040-01-01", resp.RetirementDate)
	assert.Greater(t, resp.FERSAnnual, 0.0)
	assert.Greater(t, resp.TSPBalance, 0.0)
	assert.Greater(t, resp.NetIncome, 0.0)
	assert.Equal(t, "4_percent_rule", resp.StrategyDescription)
}
