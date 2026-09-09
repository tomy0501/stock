package stock

import (
	"fmt"
	"stock/db"
	"stock/myRedis"
	"time"

	"github.com/google/uuid"
)

type StockRepository struct {
	db      db.UserDB
	myRedis myRedis.RedisPool
}

type StockDatabaseColumn struct {
	Date     string `db:"date"`
	Stock    int    `db:"stock"`
	Reserved int    `db:"reserved"`
	Stopped  int    `db:"stopped"`
}

func NewStockRepository(db db.UserDB, myRedis myRedis.RedisPool) StockHandlerInterface {
	return &StockRepository{
		db:      db,
		myRedis: myRedis,
	}
}

func (s *StockRepository) GetStockDataPerMonth(roomTypeID string, year int, month int) (*RoomStockData, error) {
	yearMonth, err := time.Parse("2006-01", fmt.Sprintf("%04d-%02d", year, month))
	if err != nil {
		return nil, fmt.Errorf("failed to parse year and month: %v", err)
	}
	firstDay := yearMonth.Format("2006-01-02")
	lastDay := yearMonth.AddDate(0, 1, -1).Format("2006-01-02")

	parsedRoomTypeID, err := uuid.Parse(roomTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room type ID: %v", err)
	}

	sql := `
		SELECT date, stock, reserved, stopped
		FROM room_stock
		WHERE room_type_id = ? AND date BETWEEN ? AND ?
		ORDER BY date ASC
	`
	var stockData []StockDatabaseColumn

	err = s.db.UserDB.Select(&stockData, sql, parsedRoomTypeID, firstDay, lastDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock data: %v", err)
	}
	var result RoomStockData
	var stockDataList []StockData
	isFirst := true
	for _, row := range stockData {
		if isFirst {
			result.RoomTypeID = roomTypeID
			isFirst = false
		}
		stockDataList = append(stockDataList, StockData{
			Date:     row.Date,
			Stock:    row.Stock,
			Reserved: row.Reserved,
			Stopped:  row.Stopped,
		})
	}
	result.StockData = stockDataList
	return &result, nil
}
