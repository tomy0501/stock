package stock

import "github.com/gin-gonic/gin"

type StockData struct {
	Date     string `json:"date" binding:"required,datetime=2006-01-02"`
	Stock    int    `json:"stock" binding:"required"`
	Reserved int    `json:"reserved" binding:"required"`
	Stopped  int    `json:"stopped" binding:"required"`
}

type RoomStockData struct {
	RoomTypeID string      `json:"room_type_id" binding:"required"`
	StockData  []StockData `json:"stock_data" binding:"required,dive"`
}

type StockHandlerInterface interface {
	GetStockDataPerMonth(roomTypeID string, year int, month int) (*RoomStockData, error)
}

type stockHandler struct {
	StockHandlerIF StockHandlerInterface
}

type StockHandleURL struct {
	RoomTypeID string `uri:"room_type_id" binding:"required"`
	Month      int    `uri:"month" binding:"required"`
	Day        int    `uri:"day" binding:"required"`
}

func NewStockHandler(stockHandlerIF StockHandlerInterface) *stockHandler {
	return &stockHandler{
		StockHandlerIF: stockHandlerIF,
	}
}

func (s *stockHandler) GetPerMonth(c *gin.Context) {
	var urlParams StockHandleURL
	if err := c.ShouldBindUri(&urlParams); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	stockData, err := s.StockHandlerIF.GetStockDataPerMonth(urlParams.RoomTypeID, urlParams.Month, urlParams.Day)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stockData)
}
