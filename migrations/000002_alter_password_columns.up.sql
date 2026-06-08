-- Migration 000002: Alter password columns to accommodate bcrypt hashes (60 chars)
-- VARCHAR(72) provides headroom for future hash algorithm changes

ALTER TABLE `tbl_account` MODIFY COLUMN `password` VARCHAR(72) NOT NULL;
ALTER TABLE `tbl_admin` MODIFY COLUMN `password` VARCHAR(72) NOT NULL;
