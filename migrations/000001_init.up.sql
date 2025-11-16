-- Создание таблиц согласно ТЗ

-- Основная таблица лотов
CREATE TABLE IF NOT EXISTS lots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  guid TEXT UNIQUE NOT NULL,
  link TEXT,
  title TEXT,
  pub_date TIMESTAMPTZ,
  dc_date TIMESTAMPTZ,
  auction_date TIMESTAMPTZ,
  cadastral_number TEXT,
  centroid GEOMETRY(POINT, 3857),
  geometry GEOMETRY(GEOMETRY, 3857),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Динамические поля из description
CREATE TABLE IF NOT EXISTS lot_fields (
  lot_id UUID REFERENCES lots(id) ON DELETE CASCADE,
  field_name TEXT NOT NULL,
  field_value TEXT,
  PRIMARY KEY (lot_id, field_name)
);

-- Данные из внешних API
CREATE TABLE IF NOT EXISTS external_data (
  lot_id UUID PRIMARY KEY REFERENCES lots(id) ON DELETE CASCADE,
  lot_info JSONB,
  lot_info_fetched_at TIMESTAMPTZ,
  lot_info_error TEXT,
  nspd_data JSONB,
  nspd_fetched_at TIMESTAMPTZ,
  nspd_error TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_lots_pub_date ON lots(pub_date);
CREATE INDEX IF NOT EXISTS idx_lots_auction_date ON lots(auction_date);
CREATE INDEX IF NOT EXISTS idx_lots_cadastral_number ON lots(cadastral_number);
CREATE INDEX IF NOT EXISTS idx_lots_guid ON lots(guid);
CREATE INDEX IF NOT EXISTS idx_lots_centroid ON lots USING GIST(centroid);
CREATE INDEX IF NOT EXISTS idx_lots_geometry ON lots USING GIST(geometry);