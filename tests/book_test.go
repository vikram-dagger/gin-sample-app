package tests

import (
	"book/controllers"
	"book/database"
	"book/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var sampleBooks = []models.Book{
	{Title: "1984", Author: "George Orwell"},
	{Title: "Brave New World", Author: "Aldous Huxley"},
	{Title: "Fahrenheit 451", Author: "Ray Bradbury"},
}

func setupTest(t *testing.T) *gin.Engine {
	database.ConnectDatabase()

	err := database.DB.Exec("TRUNCATE TABLE books RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	r := gin.Default()
	r.POST("/api/books", controllers.CreateBook)
	r.GET("/api/books", controllers.GetBooks)
	r.GET("/api/books/:id", controllers.GetBookByID)
	r.DELETE("/api/books/:id", controllers.DeleteBook)
	return r
}

func seedBooks(t *testing.T) {
	for _, b := range sampleBooks {
		if err := database.DB.Create(&b).Error; err != nil {
			t.Fatalf("Failed to seed book (%s): %v", b.Title, err)
		}
	}
}

func TestCreateBook(t *testing.T) {
	router := setupTest(t)

	jsonValue, _ := json.Marshal(sampleBooks[0])
	req, _ := http.NewRequest("POST", "/api/books", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected status 201 Created, got %d. Response body: %s", w.Code, w.Body.String())

	var resp models.Book
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Failed to unmarshal response: %v", err)

	assert.Equal(t, sampleBooks[0].Title, resp.Title)
	assert.Equal(t, sampleBooks[0].Author, resp.Author)
}

func TestGetBooks(t *testing.T) {
	router := setupTest(t)
	seedBooks(t)

	req, _ := http.NewRequest("GET", "/api/books", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())

	var books []models.Book
	err := json.Unmarshal(w.Body.Bytes(), &books)
	assert.NoError(t, err, "Failed to unmarshal books list: %v", err)
	assert.Equal(t, len(sampleBooks), len(books), "Expected %d books, got %d", len(sampleBooks), len(books))
}

func TestGetBookByID(t *testing.T) {
	router := setupTest(t)
	seedBooks(t)

	var firstBook models.Book
	if err := database.DB.First(&firstBook).Error; err != nil {
		t.Fatalf("Failed to fetch seeded book: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/books/"+strconv.Itoa(int(firstBook.ID)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())

	var fetched models.Book
	err := json.Unmarshal(w.Body.Bytes(), &fetched)
	assert.NoError(t, err, "Failed to unmarshal response: %v", err)

	assert.Equal(t, firstBook.ID, fetched.ID)
	assert.Equal(t, firstBook.Title, fetched.Title)
	assert.Equal(t, firstBook.Author, fetched.Author)
}

func TestDeleteBook(t *testing.T) {
	router := setupTest(t)
	seedBooks(t)

	var firstBook models.Book
	if err := database.DB.First(&firstBook).Error; err != nil {
		t.Fatalf("Failed to fetch seeded book: %v", err)
	}

	req, _ := http.NewRequest("DELETE", "/api/books/"+strconv.Itoa(int(firstBook.ID)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())

	// Verify the book was deleted
	var deletedBook models.Book
	err := database.DB.First(&deletedBook, firstBook.ID).Error
	assert.Error(t, err, "Book should have been deleted")

	// Try to delete non-existent book
	req, _ = http.NewRequest("DELETE", "/api/books/999999", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 Not Found for non-existent book")
}
