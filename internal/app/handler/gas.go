package handler

import (
	"net/http"
	"strconv"

	"WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllGases(ctx *gin.Context) {
	var gas []ds.Gas
	var err error

	search := ctx.Query("search")
	if search == "" {
		gas, err = h.Repository.GetAllGases()
	} else {
		gas, err = h.Repository.SearchGasesByTitle(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"gases":      gas,
		"cart_count": h.Repository.GetCartCount(),
		"search":     search,
	})
}
func (h *Handler) GetGasById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gas ID"})
		return
	}

	gas, err := h.Repository.GetGasByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.HTML(http.StatusOK, "gases.html", gin.H{
		"gas":        gas,
		"cart_count": h.Repository.GetCartCount(),
	})
}
