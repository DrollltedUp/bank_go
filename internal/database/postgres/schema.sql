-- Расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Отделения банков
CREATE TABLE IF NOT EXISTS branches (
    branch_id VARCHAR(255) PRIMARY KEY,
    bank_name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    location_type VARCHAR(50) DEFAULT 'branch',
    opening_hours TEXT,
    phone VARCHAR(50),
    windows INTEGER DEFAULT 2,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Единая очередь для отделения
CREATE TABLE IF NOT EXISTS queues (
    queue_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id VARCHAR(255) REFERENCES branches(branch_id) ON DELETE CASCADE,
    current_number INTEGER DEFAULT 0,
    tickets_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id)
);

-- Типы услуг (справочник)
CREATE TABLE IF NOT EXISTS service_types (
    service_code VARCHAR(10) PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20) DEFAULT '#4CAF50',
    is_active BOOLEAN DEFAULT true
);

-- Талоны (единая очередь)
CREATE TABLE IF NOT EXISTS tickets (
    ticket_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number VARCHAR(20) NOT NULL,
    service_code VARCHAR(10) REFERENCES service_types(service_code),
    branch_id VARCHAR(255) REFERENCES branches(branch_id) ON DELETE CASCADE,
    queue_id UUID REFERENCES queues(queue_id),
    position INTEGER NOT NULL,
    wait_time INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    called_at TIMESTAMP,
    completed_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'waiting',
    CONSTRAINT valid_status CHECK (status IN ('waiting', 'called', 'completed', 'cancelled'))
);

-- История загруженности
CREATE TABLE IF NOT EXISTS branch_load_history (
    history_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id VARCHAR(255) REFERENCES branches(branch_id) ON DELETE CASCADE,
    load_score INTEGER CHECK (load_score BETWEEN 1 AND 5),
    tickets_total INTEGER,
    windows INTEGER,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_tickets_branch_status ON tickets(branch_id, status);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets(created_at);
CREATE INDEX IF NOT EXISTS idx_branch_history_time ON branch_load_history(recorded_at);

-- Вставка типов услуг
INSERT INTO service_types (service_code, service_name, description, color) VALUES
    ('CASH', 'Кассовое обслуживание', 'Оплата счетов, переводы, валюта', '#FF6B6B'),
    ('PENSION', 'Пенсии и пособия', 'Выплата пенсий, социальные выплаты', '#4ECDC4'),
    ('DEBIT', 'Дебетовые карты', 'Оформление и перевыпуск', '#45B7D1'),
    ('CREDIT', 'Кредитные карты', 'Оформление и консультация', '#96CEB4'),
    ('MORTGAGE', 'Ипотека и кредиты', 'Оформление ипотеки, автокредитов', '#FFEAA7'),
    ('VIP', 'Премиум-обслуживание', 'VIP-клиенты', '#DDA0DD'),
    ('BUSINESS', 'Юридическим лицам', 'Расчетно-кассовое обслуживание', '#98D8C8')
ON CONFLICT (service_code) DO NOTHING;

-- Функция обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры
DROP TRIGGER IF EXISTS update_branches_updated_at ON branches;
CREATE TRIGGER update_branches_updated_at 
    BEFORE UPDATE ON branches 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_queues_updated_at ON queues;
CREATE TRIGGER update_queues_updated_at 
    BEFORE UPDATE ON queues 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();