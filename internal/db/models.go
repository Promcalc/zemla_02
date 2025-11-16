package db

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// LotDB — структура для сохранения в таблицу lots
type LotDB struct {
	ID              string    `db:"id"`
	GUID            string    `db:"guid"`
	Link            string    `db:"link"`
	Title           string    `db:"title"`
	PubDate         time.Time `db:"pub_date"`
	DCDate          time.Time `db:"dc_date"`
	AuctionDate     time.Time `db:"auction_date"`
	CadastralNumber string    `db:"cadastral_number"`
	Centroid        string    `db:"centroid"`    // WKT или NULL
	Geometry        string    `db:"geometry"`    // WKT или NULL
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// LotFieldDB — запись для таблицы lot_fields
type LotFieldDB struct {
	LotID    string `db:"lot_id"`
	FieldName string `db:"field_name"`
	FieldValue string `db:"field_value"`
}

// JSONB — обёртка для корректной работы с JSONB в PostgreSQL
type JSONB struct {
	Data interface{}
}

func (j JSONB) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// ExternalDataDB — структура для таблицы external_data
type ExternalDataDB struct {
	LotID           string     `db:"lot_id"`
	LotInfo         JSONB      `db:"lot_info"`
	LotInfoFetched  *time.Time `db:"lot_info_fetched_at"`
	LotInfoError    *string    `db:"lot_info_error"`
	NSPDData        JSONB      `db:"nspd_data"`
	NSPDFetched     *time.Time `db:"nspd_fetched_at"`
	NSPDError       *string    `db:"nspd_error"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}