ALTER TABLE payment_intents DROP COLUMN mode;
ALTER TABLE payment_intents DROP COLUMN zone_id;

ALTER TABLE accounts DROP COLUMN mode;
ALTER TABLE accounts DROP COLUMN zone_id;

ALTER TABLE transactions DROP COLUMN mode;
ALTER TABLE transactions DROP COLUMN zone_id;
