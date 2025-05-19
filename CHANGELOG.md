diff --git a/controllers/book_controller.go b/controllers/book_controller.go
index c212e0c..4bf6162 100644
--- a/controllers/book_controller.go
+++ b/controllers/book_controller.go
@@ -44,3 +44,22 @@ func GetBookByID(c *gin.Context) {
 
 	c.JSON(http.StatusOK, book)
 }
+
+func DeleteBook(c *gin.Context) {
+	var book models.Book
+	id := c.Param("id")
+
+	// First check if the book exists
+	if err := database.DB.First(&book, id).Error; err != nil {
+		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found", "details": err.Error()})
+		return
+	}
+
+	// Delete the book
+	if err := database.DB.Delete(&book).Error; err != nil {
+		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book", "details": err.Error()})
+		return
+	}
+
+	c.JSON(http.StatusOK, book)
+}
diff --git a/routes/routes.go b/routes/routes.go
index b630b08..0b17d70 100644
--- a/routes/routes.go
+++ b/routes/routes.go
@@ -12,5 +12,6 @@ func RegisterRoutes(r *gin.Engine) {
 		api.POST("/books", controllers.CreateBook)
 		api.GET("/books", controllers.GetBooks)
 		api.GET("/books/:id", controllers.GetBookByID)
+		api.DELETE("/books/:id", controllers.DeleteBook)
 	}
 }
diff --git a/tests/book_test.go b/tests/book_test.go
index a6f6780..2e961f0 100644
--- a/tests/book_test.go
+++ b/tests/book_test.go
@@ -33,6 +33,7 @@ func setupTest(t *testing.T) *gin.Engine {
 	r.POST("/api/books", controllers.CreateBook)
 	r.GET("/api/books", controllers.GetBooks)
 	r.GET("/api/books/:id", controllers.GetBookByID)
+	r.DELETE("/api/books/:id", controllers.DeleteBook)
 	return r
 }
 
@@ -103,3 +104,31 @@ func TestGetBookByID(t *testing.T) {
 	assert.Equal(t, firstBook.Title, fetched.Title)
 	assert.Equal(t, firstBook.Author, fetched.Author)
 }
+
+func TestDeleteBook(t *testing.T) {
+	router := setupTest(t)
+	seedBooks(t)
+
+	var firstBook models.Book
+	if err := database.DB.First(&firstBook).Error; err != nil {
+		t.Fatalf("Failed to fetch seeded book: %v", err)
+	}
+
+	req, _ := http.NewRequest("DELETE", "/api/books/"+strconv.Itoa(int(firstBook.ID)), nil)
+	w := httptest.NewRecorder()
+	router.ServeHTTP(w, req)
+
+	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
+
+	// Verify the book was deleted
+	var deletedBook models.Book
+	err := database.DB.First(&deletedBook, firstBook.ID).Error
+	assert.Error(t, err, "Book should have been deleted")
+
+	// Try to delete non-existent book
+	req, _ = http.NewRequest("DELETE", "/api/books/999999", nil)
+	w = httptest.NewRecorder()
+	router.ServeHTTP(w, req)
+
+	assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 Not Found for non-existent book")
+}
