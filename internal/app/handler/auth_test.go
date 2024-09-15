package handler

import (
	"api/internal/config"
	mockmailer "api/internal/mailer/mock"
	"api/internal/repository"
	repoerr "api/internal/repository/errors"
	mocktoken "api/internal/token/mock"
	"api/pkg/sha256"
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	tokenSecret := "some-supa-secret-characters"
	c := config.Config{Token: config.Token{Secret: tokenSecret}}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))
	handler := New(&c, repo, mockmailer.New(), mocktoken.New(c.Token))

	tests := []table{
		{
			name: "ok",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_private", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "", sha256.String("testword"), false, false, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING *").
					WithArgs("john.doe@example.com", "johndoe", sha256.String("testword")).WillReturnRows(rows)
			},

			request: request{
				body: `{"email":"john.doe@example.com","username":"johndoe","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusCreated,
				body:   `{"id":"USER_ID","email":"john.doe@example.com","username":"johndoe","display_name":"","is_private":false,"is_confirmed":false}`,
			},
		},
		{
			name: "invalid request body",

			request: request{
				body: `{"some":"invalid","request":"structure"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid request body"}`,
			},
		},
		{
			name: "invalid email format",

			request: request{
				body: `{"email":"incorrect-email","username":"John Doe","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid email format"}`,
			},
		},
		{
			name: "usernamename is too short",

			request: request{
				body: `{"email":"john.doe@example.com","username":"wort","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"username is too short"}`,
			},
		},
		{
			name: "password is too long",

			request: request{
				body: `{"email":"john.doe@example.com","username":"John Doe","password":"very-looooooooooooooooooooooooooooooooooooooooooong-password"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"password is too long"}`,
			},
		},
		{
			name: "user already exists",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING *").
					WithArgs("john.doe@example.com", "John Doe", sha256.String("testword")).WillReturnError(repoerr.ErrUserAlreadyExists)
			},

			request: request{
				body: `{"email":"john.doe@example.com","username":"John Doe","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusConflict,
				body:   `{"message":"user already exists"}`,
			},
		},
		{
			name: "repository error",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING *").
					WithArgs("john.doe@example.com", "John Doe", sha256.String("testword")).WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"email":"john.doe@example.com","username":"John Doe","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, TemplateTestHandler(tc, mock, http.MethodPost, "/api/auth/account", handler.CreateAccount))
	}
}

func TestCreateSession(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	tokenSecret := "some-supa-secret-characters"
	c := config.Config{Token: config.Token{Secret: tokenSecret}}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))
	handler := New(&c, repo, mockmailer.New(), mocktoken.New(c.Token))

	tests := []table{
		{
			name: "ok",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "password_hash", "is_confirmed", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", sha256.String("testword"), true, time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
					WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnRows(rows)
			},

			request: request{
				body: `{"email":"john.doe@example.com","password":"testword"}`,
			},

			expect: expect{
				status:     http.StatusOK,
				bodyFields: []string{"token"},
			},
		},
		{
			name: "invalid request body",

			request: request{
				body: `{"some":"invalid","body":"poo"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid request body"}`,
			},
		},
		{
			name: "user not found",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
					WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnError(repoerr.ErrUserNotFound)
			},

			request: request{
				body: `{"email":"john.doe@example.com","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusUnauthorized,
				body:   `{"message":"user not found"}`,
			},
		},
		{
			name: "repository error",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").
					WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"email":"john.doe@example.com","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
		{
			name: "user not confirmed",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "", sha256.String("testword"), false, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE email = $1 AND password_hash = $2").WithArgs("john.doe@example.com", sha256.String("testword")).WillReturnRows(rows)
			},

			request: request{
				body: `{"email":"john.doe@example.com","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusForbidden,
				body:   `{"message":"email confirmation needed"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, TemplateTestHandler(tc, mock, http.MethodPost, "/api/auth/session", handler.CreateSession))
	}
}

func TestResetPassword(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	tokenSecret := "some-supa-secret-characters"
	c := config.Config{Token: config.Token{Secret: tokenSecret}}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))
	handler := New(&c, repo, mockmailer.New(), mocktoken.New(c.Token))

	tests := []table{
		{
			name: "ok",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "John Doe", sha256.String("testword"), true, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE email = $1").WithArgs("john.doe@example.com").WillReturnRows(rows)

				rows = sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", false, time.Now().Add(5*time.Minute).Truncate(time.Minute), time.Now())

				mock.ExpectQuery("INSERT INTO requests (email, token, expires_at) VALUES ($1, $2, $3) RETURNING *").WithArgs("john.doe@example.com", "LONGTOKEN", time.Now().Add(5*time.Minute).Truncate(time.Minute)).WillReturnRows(rows)
			},

			request: request{
				body: `{"email":"john.doe@example.com"}`,
			},

			expect: expect{
				status: http.StatusOK,
			},
		},
		{
			name: "invalid request body",

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid request body"}`,
			},
		},
		{
			name: "user not found",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE email = $1").WithArgs("john.doe@example.com").WillReturnError(repoerr.ErrUserNotFound)
			},

			request: request{
				body: `{"email":"john.doe@example.com"}`,
			},

			expect: expect{
				status: http.StatusNotFound,
				body:   `{"message":"user not found"}`,
			},
		},
		{
			name: "repository error",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE email = $1").WithArgs("john.doe@example.com").WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"email":"john.doe@example.com"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
		{
			name: "repository error",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "John Doe", sha256.String("testword"), true, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE email = $1").WithArgs("john.doe@example.com").WillReturnRows(rows)

				mock.ExpectQuery("INSERT INTO requests (email, token, expires_at) VALUES ($1, $2, $3) RETURNING *").WithArgs("john.doe@example.com", "LONGTOKEN", time.Now().Add(5*time.Minute).Truncate(time.Minute)).WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"email":"john.doe@example.com"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, TemplateTestHandler(tc, mock, http.MethodPost, "/api/auth/password/reset", handler.ResetPassword))
	}
}

func TestUpdatePassword(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	c := config.Config{}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))
	handler := New(&c, repo, mockmailer.New(), mocktoken.New(c.Token))

	tests := []table{
		{
			name: "ok",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", false, time.Now().Add(5*time.Minute), time.Now())

				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET password_hash = $1 WHERE email = $2").
					WithArgs(sha256.String("testword"), "john.doe@example.com").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("UPDATE requests SET is_used = true WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnResult(sqlmock.NewResult(1, 1))
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusOK,
			},
		},
		{
			name: "invalid request body",

			request: request{
				body: `{"some":"invalid","request":"body"}`,
			},

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid request body"}`,
			},
		},
		{
			name: "token doesn't exists",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnError(repoerr.ErrRequestNotFound)
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusNotFound,
				body:   `{"message":"password reset request not found"}`,
			},
		},
		{
			name: "repository error on getting token",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
		{
			name: "reset password request expired",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", false, time.Now().Add(-5*time.Minute), time.Now())

				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnRows(rows)
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusForbidden,
				body:   `{"message":"recovery token expired"}`,
			},
		},
		{
			name: "reset password request already used",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", true, time.Now().Add(5*time.Minute), time.Now())

				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnRows(rows)
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusForbidden,
				body:   `{"message":"this recovery token has been used"}`,
			},
		},
		{
			name: "repository error on updating password",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", false, time.Now().Add(5*time.Minute), time.Now())

				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET password_hash = $1 WHERE email = $2").
					WithArgs(sha256.String("testword"), "john.doe@example.com").
					WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
		{
			name: "can't mark request as used",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "token", "is_used", "expires_at", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "LONGTOKEN", false, time.Now().Add(5*time.Minute), time.Now())

				mock.ExpectQuery("SELECT * FROM requests WHERE token = $1").
					WithArgs("LONGTOKEN").WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET password_hash = $1 WHERE email = $2").
					WithArgs(sha256.String("testword"), "john.doe@example.com").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("UPDATE requests SET is_used = true WHERE token = $1").
					WithArgs("LONGTOKEN").
					WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"token":"LONGTOKEN","password":"testword"}`,
			},

			expect: expect{
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, TemplateTestHandler(tc, mock, http.MethodPatch, "/api/auth/password", handler.UpdatePassword))
	}
}

func TestConfirmAccount(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("err not expected: %v\n", err)
	}

	c := config.Config{}
	repo := repository.New(sqlx.NewDb(db, "sqlmock"))
	handler := New(&c, repo, mockmailer.New(), mocktoken.New(c.Token))

	tests := []table{
		{
			name: "ok",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "John Doe", sha256.String("testword"), true, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE confirmation_token = $1").
					WithArgs("CONFIRMATION_TOKEN").
					WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET is_confirmed = true WHERE email = $1 AND confirmation_token = $2").
					WithArgs("john.doe@example.com", "CONFIRMATION_TOKEN").
					WillReturnResult(driver.RowsAffected(1))
			},

			request: request{
				body: `{"token":"CONFIRMATION_TOKEN"}`,
			},

			expect: expect{
				status: http.StatusOK,
			},
		},
		{
			name: "invalid request body",

			expect: expect{
				status: http.StatusBadRequest,
				body:   `{"message":"invalid request body"}`,
			},
		},
		{
			name: "request not found",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE confirmation_token = $1").
					WithArgs("CONFIRMATION_TOKEN").
					WillReturnError(repoerr.ErrUserNotFound)
			},

			request: request{
				body: `{"token":"CONFIRMATION_TOKEN"}`,
			},

			expect: expect{
				status: http.StatusNotFound,
				body:   `{"message":"user not found"}`,
			},
		},
		{
			name: "repository error",

			repo: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM users WHERE confirmation_token = $1").
					WithArgs("CONFIRMATION_TOKEN").
					WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"token":"CONFIRMATION_TOKEN"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
		{
			name: "repository error on confirming",

			repo: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username", "display_name", "password_hash", "is_confirmed", "confirmation_token", "created_at"}).
					AddRow("USER_ID", "john.doe@example.com", "johndoe", "John Doe", sha256.String("testword"), true, "CONFIRMATION_TOKEN", time.Now())

				mock.ExpectQuery("SELECT * FROM users WHERE confirmation_token = $1").
					WithArgs("CONFIRMATION_TOKEN").
					WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET is_confirmed = true WHERE email = $1 AND confirmation_token = $2").
					WithArgs("john.doe@example.com", "CONFIRMATION_TOKEN").
					WillReturnError(errors.New("repo: Some repository error"))
			},

			request: request{
				body: `{"token":"CONFIRMATION_TOKEN"}`,
			},

			expect: expect{
				status: http.StatusInternalServerError,
				body:   `{"message":"internal server error"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, TemplateTestHandler(tc, mock, http.MethodPost, "/api/auth/account/confirm", handler.ConfirmAccount))
	}
}
