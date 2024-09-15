package handler

import (
	"api/internal/app/handler/response"
	"api/internal/app/handler/response/responsebody"
	"api/internal/lib/logger/sl"
	repoerr "api/internal/repository/errors"
	"api/pkg/requestid"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary      Get public information about user by username
// @Description  returns an user's information and week activity history
// @Tags         user
// @Produce      json
// @Param        username          path string true "Username"
// @Success      200 {object}      responsebody.PublicProfile
// @Success      200 {object}      responsebody.PrivateProfile
// @Failure      404 {object}      responsebody.Message
// @Router       /user/{username}  [get]
func (h *Handler) GetUserByUsername(c *gin.Context) {
	log := slog.With(
		slog.String("op", "handler.GetUserByUsername"),
		slog.String("request_id", requestid.Get(c)),
	)

	username := c.Param("username")

	user, err := h.repository.User.GetByUsername(c, username)
	if errors.Is(err, repoerr.ErrUserNotFound) {
		log.Debug("user not found")
		response.WithMessage(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error("could not get user by username", sl.Err(err), slog.String("username", username))
		response.InternalServerError(c)
		return
	}

	if user.IsPrivate {
		c.JSON(http.StatusOK, responsebody.PrivateProfile{
			ID:        user.ID,
			Username:  user.Username,
			IsPrivate: user.IsPrivate,
		})
		return
	}

	workouts, err := h.repository.Workout.GetUserWorkouts(c, user.ID, time.Now().Add(-6*24*time.Hour).Truncate(24*time.Hour), time.Now().Add(time.Hour).Truncate(time.Hour))
	if err != nil {
		log.Error("could not get workouts", sl.Err(err))
		response.InternalServerError(c)
		return
	}

	activity := make([]responsebody.Workout, 0)

	fmt.Println("workouts ", workouts)
	fmt.Println("activity ", activity)

	for _, workout := range workouts {
		activity = append(activity, responsebody.Workout{
			ID:       workout.ID,
			Date:     workout.Date.Format("02-01-2006"),
			Duration: workout.Duration,
			Kind:     workout.Kind,
		})
	}

	c.JSON(http.StatusOK, responsebody.PublicProfile{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		IsPrivate:    user.IsPrivate,
		WeekActivity: activity,
	})
}
