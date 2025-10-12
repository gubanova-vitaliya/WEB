package handler

import (
	"net/http"

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
		"data":   gas,
		"search": search,
	})
}
