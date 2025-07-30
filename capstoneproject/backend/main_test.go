package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	router := setupRouter(nil)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestGetGoals(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "goal_name"}).
		AddRow(1, "Test Goal")
	mock.ExpectQuery(SelectAllGoalsQuery).WillReturnRows(rows)

	router := setupRouter(db)

	req, _ := http.NewRequest("GET", PathGoals, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, true, jsonResponse["success"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGoalsDbError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(SelectAllGoalsQuery).WillReturnError(sql.ErrConnDone)

	router := setupRouter(db)

	req, _ := http.NewRequest("GET", PathGoals, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGoals_RowScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Intentionally provide wrong column type to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "goal_name"}).
		AddRow("wrong_type", "Test Goal")

	mock.ExpectQuery(SelectAllGoalsQuery).WillReturnRows(rows)

	router := setupRouter(db)

	req, _ := http.NewRequest("GET", PathGoals, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddGoal(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO goals").
		WithArgs("New Goal").
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := setupRouter(db)
	payload := []byte(`{"goal_name": "New Goal"}`)
	req, _ := http.NewRequest("POST", PathGoals, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, true, jsonResponse["success"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddGoal_DbError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO goals").
		WithArgs("Fail Goal").
		WillReturnError(sql.ErrConnDone)

	router := setupRouter(db)
	payload := []byte(`{"goal_name": "Fail Goal"}`)
	req, _ := http.NewRequest("POST", PathGoals, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddGoal_LastInsertIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO goals").
		WithArgs("New Goal").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("LastInsertId error")))

	router := setupRouter(db)
	payload := []byte(`{"goal_name": "New Goal"}`)
	req, _ := http.NewRequest("POST", PathGoals, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddGoal_InvalidPayload(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := setupRouter(db)
	payload := []byte(`{"title": "Missing goal_name field"}`)
	req, _ := http.NewRequest("POST", PathGoals, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteGoal(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM goals").WithArgs("1").WillReturnResult(sqlmock.NewResult(1, 1))

	router := setupRouter(db)
	req, _ := http.NewRequest("DELETE", "/goals/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteGoal_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM goals").WithArgs("999").WillReturnResult(sqlmock.NewResult(0, 0))

	router := setupRouter(db)
	req, _ := http.NewRequest("DELETE", "/goals/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}
