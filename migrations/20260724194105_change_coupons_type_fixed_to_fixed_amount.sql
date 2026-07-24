-- +goose Up
-- Update data existing (kalau ada baris pakai 'fixed')
UPDATE coupons SET type = 'fixed_amount' WHERE type = 'fixed';

-- Drop constraint lama
ALTER TABLE coupons DROP CONSTRAINT coupons_type_check;

-- Tambah constraint baru dengan nilai baru
ALTER TABLE coupons
    ADD CONSTRAINT coupons_type_check
    CHECK (type IN ('percentage', 'fixed_amount'));


-- +goose Down
-- Kembalikan data
UPDATE coupons SET type = 'fixed' WHERE type = 'fixed_amount';

-- Drop constraint baru
ALTER TABLE coupons DROP CONSTRAINT coupons_type_check;

-- Kembalikan constraint lama
ALTER TABLE coupons
    ADD CONSTRAINT coupons_type_check
    CHECK (type IN ('percentage', 'fixed'));

