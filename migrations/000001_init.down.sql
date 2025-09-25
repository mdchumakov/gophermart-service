-- Удаление триггеров
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление триггерной функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление представления
DROP VIEW IF EXISTS user_balance;

-- Удаление индексов
DROP INDEX IF EXISTS idx_users_login;
DROP INDEX IF EXISTS idx_withdrawals_processed_at;
DROP INDEX IF EXISTS idx_withdrawals_user_id;
DROP INDEX IF EXISTS idx_orders_uploaded_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;

-- Удаление таблиц (в правильном порядке из-за внешних ключей)
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;



