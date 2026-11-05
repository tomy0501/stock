package stock

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StockData struct {
	Date     string `json:"date" binding:"required,datetime=2006/01/02"`
	Stock    int    `json:"stock" binding:"required"`
	Reserved int    `json:"reserved" binding:"required"`
	Stopped  int    `json:"stopSell" binding:"required"`
}

type RoomStockData struct {
	RoomTypeID string      `json:"room_type_id" binding:"required"`
	StockData  []StockData `json:"stock_data" binding:"required,dive"`
}

type RoomStockProperty string

const (
	RoomStockPropertyStock    RoomStockProperty = "stock"
	RoomStockPropertyReserved RoomStockProperty = "reserved"
	RoomStockPropertyStopped  RoomStockProperty = "stopSell"
)

type StockHandlerInterface interface {
	GetStockDataPerMonth(roomTypeID string, year int, month int, facilityID string) (*RoomStockData, error)
	PutStockData(roomTypeID string, date string, facilityID string, value int) error
	PutReservedData(roomTypeID string, date string, facilityID string, value int) error
	PutStopSellData(roomTypeID string, date string, facilityID string, value int) error
}

type stockHandler struct {
	StockHandlerIF StockHandlerInterface
}

type StockHandleURL struct {
	FacilityID string `uri:"facility_id" binding:"required"`
	RoomTypeID string `uri:"room_type_id" binding:"required"`
	Month      int    `uri:"month" binding:"required"`
	Day        int    `uri:"day" binding:"required"`
}

func NewStockHandler(stockHandlerIF StockHandlerInterface) *stockHandler {
	return &stockHandler{
		StockHandlerIF: stockHandlerIF,
	}
}

type UpdateStockRequest struct {
	RoomTypeID string `uri:"room_type_id" binding:"required"`
	Year       int    `uri:"year" binding:"required"`
	Month      int    `uri:"month" binding:"required"`
	Day        int    `uri:"day" binding:"required"`
	FacilityID string `uri:"facility_id" binding:"required"`
	Value      int    `uri:"value" binding:"required"`
}

type UpdateStockJsonRequest struct {
	RoomTypeID string `json:"room_type_id" binding:"required"`
	Date       string `json:"date" binding:"required,datetime=2006/01/02"`
	Stock      *int   `json:"stock" binding:"gte=0"`
	Reserved   *int   `json:"reserved" binding:"gte=0"`
	Stopped    *int   `json:"stopSell" binding:"gte=0"`
}

type UpdateStockUri struct {
	FacilityID string `uri:"facility_id" binding:"required"`
}

func (s *stockHandler) GetPerMonth(c *gin.Context) {
	var urlParams StockHandleURL
	if err := c.ShouldBindUri(&urlParams); err != nil {
		slog.Error("failed to bind URL parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid URL parameters"})
		return
	}
	useID := c.GetString("userID")
	if useID == "" {
		slog.Error("not exists userID")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid session status"})
		return
	}

	stockData, err := s.StockHandlerIF.GetStockDataPerMonth(urlParams.RoomTypeID, urlParams.Month, urlParams.Day, urlParams.FacilityID)
	if err != nil {
		slog.Error("failed to get stock data", "error", err)
		c.JSON(500, gin.H{"message": "failed to get stock data"})
		return
	}

	c.JSON(200, gin.H{"message": "success", "roomStocks": stockData})
}

func (s *stockHandler) UpdateStock(c *gin.Context) {
	var urlParams UpdateStockRequest
	if err := c.ShouldBindUri(&urlParams); err != nil {
		slog.Error("failed to bind URL parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid URL parameters"})
		return
	}
	userID := c.GetString("userID")
	if userID == "" {
		slog.Error("not exists userID")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid session status"})
		return
	}
	date := fmt.Sprintf("%04d-%02d-%02d", urlParams.Year, urlParams.Month, urlParams.Day)
	err := s.StockHandlerIF.PutStockData(urlParams.RoomTypeID, date, urlParams.FacilityID, urlParams.Value)

	if err != nil {
		slog.Error("failed to update stock data", "error", err)
		c.JSON(500, gin.H{"message": "failed to update stock data"})
		return
	}

	c.JSON(200, gin.H{"message": "success"})
}

func (s *stockHandler) UpdateReserved(c *gin.Context) {
	var urlParams UpdateStockRequest
	if err := c.ShouldBindUri(&urlParams); err != nil {
		slog.Error("failed to bind URL parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid URL parameters"})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		slog.Error("not exists userID")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid session status"})
		return
	}

	date := fmt.Sprintf("%04d-%02d-%02d", urlParams.Year, urlParams.Month, urlParams.Day)
	err := s.StockHandlerIF.PutReservedData(urlParams.RoomTypeID, date, urlParams.FacilityID, urlParams.Value)
	if err != nil {
		slog.Error("failed to update reserved data", "error", err)
		c.JSON(500, gin.H{"message": "failed to update reserved data"})
		return
	}

	c.JSON(200, gin.H{"message": "success"})
}

func (s *stockHandler) UpdateStopSell(c *gin.Context) {
	var urlParams UpdateStockRequest
	if err := c.ShouldBindUri(&urlParams); err != nil {
		slog.Error("failed to bind URL parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid URL parameters"})
		return
	}
	var userID = c.GetString("userID")
	if userID == "" {
		slog.Error("not exists userID")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid session status"})
		return
	}
	date := fmt.Sprintf("%04d-%02d-%02d", urlParams.Year, urlParams.Month, urlParams.Day)

	err := s.StockHandlerIF.PutStopSellData(urlParams.RoomTypeID, date, urlParams.FacilityID, urlParams.Value)
	if err != nil {
		slog.Error("failed to update stop sell data", "error", err)
		c.JSON(500, gin.H{"message": "failed to update stop sell data"})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func (s *stockHandler) UpdateStockJson(c *gin.Context) {
	var urlParams UpdateStockUri
	if err := c.ShouldBindUri(&urlParams); err != nil {
		slog.Error("failed to bind URL parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid URL parameters"})
		return
	}

	var jsonParams UpdateStockJsonRequest
	if err := c.ShouldBindJSON(&jsonParams); err != nil {
		slog.Error("failed to bind JSON parameters", "error", err)
		c.JSON(400, gin.H{"message": "invalid JSON parameters"})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		slog.Error("not exists userID")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid session status"})
		return
	}

	if jsonParams.Stock != nil {
		err := s.StockHandlerIF.PutStockData(jsonParams.RoomTypeID, jsonParams.Date, urlParams.FacilityID, *jsonParams.Stock)
		if err != nil {
			slog.Error("failed to update stock data", "error", err)
			c.JSON(500, gin.H{"message": "failed to update stock data"})
			return
		}
	}
	if jsonParams.Reserved != nil {
		err := s.StockHandlerIF.PutReservedData(jsonParams.RoomTypeID, jsonParams.Date, urlParams.FacilityID, *jsonParams.Reserved)
		if err != nil {
			slog.Error("failed to update reserved data", "error", err)
			c.JSON(500, gin.H{"message": "failed to update reserved data"})
			return
		}
	}
	if jsonParams.Stopped != nil {
		err := s.StockHandlerIF.PutStopSellData(jsonParams.RoomTypeID, jsonParams.Date, urlParams.FacilityID, *jsonParams.Stopped)
		if err != nil {
			slog.Error("failed to update stop sell data", "error", err)
			c.JSON(500, gin.H{"message": "failed to update stop sell data"})
			return
		}
	}
}
