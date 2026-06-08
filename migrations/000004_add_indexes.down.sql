-- Migration 000004 (down): Drop all added indexes

DROP INDEX `idx_transaction_account_no` ON `tbl_transaction`;
DROP INDEX `idx_transaction_trans_date` ON `tbl_transaction`;
DROP INDEX `idx_login_history_account_no` ON `tbl_login_history`;
DROP INDEX `idx_feedback_account_no` ON `tbl_feedback`;
