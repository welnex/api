package handler

import (
	"api/internal/app/handler/response/responsebody"
	"api/internal/config"
	"api/internal/repository"
	repoerr "api/internal/repository/errors"
	"api/pkg/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	c := config.Config{}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))

	t.Run("OK", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at"}).
			AddRow("69", "john.doe@example.com", "John Doe", sha256.String("testword"), time.Now())

		mock.ExpectQuery("INSERT INTO users (email, name, password_hash) values ($1, $2, $3) RETURNING *").
			WithArgs("john.doe@example.com", "John Doe", sha256.String("testword")).WillReturnRows(rows)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"john.doe@example.com","name":"John Doe","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusCreated
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"id":"69","email":"john.doe@example.com","name":"John Doe"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"some":"incorrect","fields":"for","request":"body"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusBadRequest
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid request body"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Invalid email", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"incorrect-email","name":"John Doe","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusBadRequest
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid email format"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Name is too long", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"john.doe@example.com","name":"very-looooooooooooooooooooooooooooooooooooooooooong-name","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusBadRequest
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"name is too long"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Password is too long", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"john.doe@example.com","name":"John Doe","password":"very-looooooooooooooooooooooooooooooooooooooong-password"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusBadRequest
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"password is too long"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("User already exists", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO users (email, name, password_hash) values ($1, $2, $3) RETURNING *").
			WithArgs("john.doe@example.com", "John Doe", sha256.String("testword")).WillReturnError(repoerr.ErrUserAlreadyExists)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"john.doe@example.com","name":"John Doe","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusConflict
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"user already exists"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Repository error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO users (email, name, password_hash) values ($1, $2, $3) RETURNING *").
			WithArgs("john.doe@example.com", "John Doe", sha256.String("testword")).WillReturnError(errors.New("repo: Some repository error"))

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/register", handler.Register)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/register",
			strings.NewReader(`{"email":"john.doe@example.com","name":"John Doe","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var body responsebody.User

		json.Unmarshal(w.Body.Bytes(), &body)

		expectedStatus := http.StatusInternalServerError
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v", status, expectedStatus)
		}

		expectedBody := `{"message":"can't register"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})
}

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	tokenSecret := "some-supa-secret-characters"
	c := config.Config{Token: config.Token{Secret: tokenSecret}}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))

	t.Run("OK", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at"}).
			AddRow("69", "john.doe@example.com", "John Doe", sha256.String("testword"), time.Now())

		mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
			WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnRows(rows)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/login", handler.Login)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"email":"john.doe@example.com","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusOK
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		var body responsebody.Token
		err = json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatalf("can't unmarshall response body: %v\n", err)
		}

		if body.Token == "" {
			t.Fatal("token should not be empty")
		}

		_, err = jwt.Parse(body.Token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(tokenSecret), nil
		})
		if err != nil {
			t.Fatalf("jsonwebtoken is invalid: %v\n", err)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/login", handler.Login)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"some":"invalid","body":"poo"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusBadRequest
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid request body"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
			WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnError(repoerr.ErrUserNotFound)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/login", handler.Login)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"email":"john.doe@example.com","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusNotFound
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"user not found"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Repository error", func(t *testing.T) {
		mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
			WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnError(errors.New("repo: Some repository error"))

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.POST("/api/auth/login", handler.Login)

		req, err := http.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"email":"john.doe@example.com","password":"testword"}`))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusInternalServerError
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"can't login"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})
}

func TestMe(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	tokenSecret := "some-supa-secret-characters"
	c := config.Config{Token: config.Token{Secret: tokenSecret}}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))

	t.Run("OK", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at"}).
			AddRow("69", "john.doe@example.com", "John Doe", sha256.String("testword"), time.Now())

		mock.ExpectQuery("SELECT * FROM users WHERE id = $1").
			WithArgs("69").WillReturnRows(rows)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat": time.Now().Unix(),
			"id":  "69",
		})

		tokenString, err := token.SignedString([]byte(c.Token.Secret))
		if err != nil {
			t.Fatalf("err not expected while signing jsonwebtoken: %v\n", err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusOK
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		var body responsebody.User
		err = json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatalf("can't unmarshall response body: %v\n", err)
		}

		if body.ID != "69" {
			t.Fatalf("unexpected user id: got %v, want %v\n", body.ID, "69")
		}
		if body.Email != "john.doe@example.com" {
			t.Fatalf("unexpected user email: got %v, want %v\n", body.ID, "john.dor@example.com")
		}
		if body.Name != "John Doe" {
			t.Fatalf("unexpected user name: got %v, want %v\n", body.ID, "John Doe")
		}
	})

	t.Run("Empty header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", fmt.Sprintf(""))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusUnauthorized
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid authorization header"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Invalid token type", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bot <token>"))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusUnauthorized
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid authorization token type"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Incorrect JWT signing method", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "some-incorrect-jwonwebtoken"))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusUnauthorized
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid authorization token"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT * FROM users WHERE id = $1").
			WithArgs("69").WillReturnError(repoerr.ErrUserNotFound)

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat": time.Now().Unix(),
			"id":  "69",
		})

		tokenString, err := token.SignedString([]byte(c.Token.Secret))
		if err != nil {
			t.Fatalf("err not expected while signing jsonwebtoken: %v\n", err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusUnauthorized
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"invalid authorization token"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})

	t.Run("Repository error", func(t *testing.T) {
		mock.ExpectQuery("SELECT * FROM users WHERE id = $1").
			WithArgs("69").WillReturnError(errors.New("repo: Some repository error"))

		gin.SetMode(gin.TestMode)
		r := gin.Default()

		handler := New(&c, repo)

		r.GET("/api/auth/me", handler.Me)

		req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if err != nil {
			t.Fatal(err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat": time.Now().Unix(),
			"id":  "69",
		})

		tokenString, err := token.SignedString([]byte(c.Token.Secret))
		if err != nil {
			t.Fatalf("err not expected while signing jsonwebtoken: %v\n", err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedStatus := http.StatusInternalServerError
		if status := w.Code; status != expectedStatus {
			t.Fatalf("handler returned wrong status code: got %v, want %v\n", status, expectedStatus)
		}

		expectedBody := `{"message":"can't get me"}`
		if w.Body.String() != expectedBody {
			t.Fatalf("handler returned unexpected body: got %v, want %v", w.Body.String(), expectedBody)
		}
	})
}
