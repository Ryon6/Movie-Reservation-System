package handlers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getIDFromPath(ctx *gin.Context) (uint, error) {
	idStr, exists := ctx.Params.Get("id")
	if !exists {
		return 0, fmt.Errorf("id not found in params")
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %w", err)
	}
	return uint(id), nil
}
