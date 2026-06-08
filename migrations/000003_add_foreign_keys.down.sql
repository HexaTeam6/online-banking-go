-- Migration 000003 (down): Drop all foreign key constraints

ALTER TABLE `tbl_account_type` DROP FOREIGN KEY `fk_account_type_account`;
ALTER TABLE `tbl_address` DROP FOREIGN KEY `fk_address_account`;
ALTER TABLE `tbl_balance` DROP FOREIGN KEY `fk_balance_account`;
ALTER TABLE `tbl_customer` DROP FOREIGN KEY `fk_customer_account`;
ALTER TABLE `tbl_feedback` DROP FOREIGN KEY `fk_feedback_account`;
ALTER TABLE `tbl_login_history` DROP FOREIGN KEY `fk_login_history_account`;
ALTER TABLE `tbl_requests` DROP FOREIGN KEY `fk_requests_account`;
ALTER TABLE `tbl_transaction` DROP FOREIGN KEY `fk_transaction_account`;
