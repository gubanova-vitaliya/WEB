package handler

import (
	"WEB/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetGases(c *gin.Context) {
	query := c.Query("query")
	var gases []repository.Gas
	var err error

	if query != "" {
		gases, err = h.repo.GetGasesByTitle(query)
	} else {
		gases, err = h.repo.GetGases()
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"error": "Ошибка загрузки газов",
			"query": query,
		})
		return
	}

	journalCount := h.repo.GetJournalCount()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"gases":        gases,
		"query":        query,
		"journalCount": journalCount,
	})
}

func (h *Handler) GetGas(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "gases.html", gin.H{
			"error": "Неверный ID газа",
		})
		return
	}

	gas, err := h.repo.GetGas(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "gases.html", gin.H{
			"error": "Газ не найден",
		})
		return
	}

	c.HTML(http.StatusOK, "gases.html", gin.H{
		"gase": gas,
	})
}

func (h *Handler) GetCart(c *gin.Context) {
	journal := h.repo.GetJournal()

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"calculations": journal.Calculations,
		"journalCount": h.repo.GetJournalCount(),
	})
}
