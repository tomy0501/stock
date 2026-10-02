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
	FacilityID string `db:"facility_id"`
	Date       string `db:"stock_date"`
	Stock      int    `db:"stock_quantity"`
	Reserved   int    `db:"reservation_quantity"`
	Stopped    int    `db:"stop_sale_quantity"`
}

func NewStockRepository(db db.UserDB, myRedis myRedis.RedisPool) StockHandlerInterface {
	return &StockRepository{
		db:      db,
		myRedis: myRedis,
	}
}

func (s *StockRepository) GetStockDataPerMonth(roomTypeID string, year int, month int, facilityID string) (*RoomStockData, error) {
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
	parsedFacilityID, err := uuid.Parse(facilityID)
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT facility_id,room_type_id,stock_date, stock_quantity, reservation_quantity, stop_sale_quantity
		FROM stock_room
		WHERE room_type_id = ? AND stock_date BETWEEN ? AND ? AND facility_id = ?
		ORDER BY stock_date ASC
	`
	var stockData []StockDatabaseColumn

	err = s.db.UserDB.Select(&stockData, sql, parsedRoomTypeID[:], firstDay, lastDay, parsedFacilityID[:])
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
