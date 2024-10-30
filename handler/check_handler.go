package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/m3talux/goauth/config"
	"github.com/m3talux/goauth/model"
	"github.com/m3talux/goauth/mongo"
	"github.com/rs/zerolog/log"
)

// CheckHandler exposes the check functions: health check and ready check.
type CheckHandler struct{}

// Alive handler is used to check whether the API is reachable.
func (cc *CheckHandler) Alive(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

// Ready handler is used to check whether the module is ready to work.
func (cc *CheckHandler) Ready(c *gin.Context) {
	errs := make([]error, 0)
	errs = append(errs, config.Check()...)
	errs = append(errs, mongo.Check()...)

	if len(errs) > 0 {
		log.Error().Interface("errors", errs).Msgf("%s is not ready", config.AppName())

		response := model.NewAPIResponseError(http.StatusServiceUnavailable, fmt.Sprintf("%v", errs))

		c.AbortWithStatusJSON(response.HTTPStatus(), response)

		return
	}

	c.JSON(http.StatusOK, "OK")
}

func NewCheckHandler() *CheckHandler {
	return &CheckHandler{}
}
