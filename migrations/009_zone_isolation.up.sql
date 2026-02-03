-- Payments table
ALTER TABLE payment_intents ADD COLUMN zone_id TEXT;
ALTER TABLE payment_intents ADD COLUMN mode TEXT CHECK (mode IN ('test', 'live'));
CREATE INDEX idx_payment_intents_zone_id ON payment_intents(zone_id);

-- Ledger tables
ALTER TABLE accounts ADD COLUMN zone_id TEXT;
ALTER TABLE accounts ADD COLUMN mode TEXT CHECK (mode IN ('test', 'live'));
CREATE INDEX idx_accounts_zone_id ON accounts(zone_id);

ALTER TABLE transactions ADD COLUMN zone_id TEXT;
ALTER TABLE transactions ADD COLUMN mode TEXT CHECK (mode IN ('test', 'live'));
CREATE INDEX idx_transactions_zone_id ON transactions(zone_id);
