-- Migration 000002 (down): Revert password columns to original CHAR(25)
-- WARNING: This will truncate any bcrypt hashes stored in these columns

ALTER TABLE `tbl_account` MODIFY COLUMN `password` CHAR(25) NOT NULL;
ALTER TABLE `tbl_admin` MODIFY COLUMN `password` CHAR(25) NOT NULL;
