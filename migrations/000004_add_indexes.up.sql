-- Migration 000004: Add indexes for frequently queried columns
-- These indexes improve query performance for transaction lookups,
-- login history filtering, and feedback retrieval

CREATE INDEX `idx_transaction_account_no` ON `tbl_transaction`(`account_no`);
CREATE INDEX `idx_transaction_trans_date` ON `tbl_transaction`(`trans_date`);
CREATE INDEX `idx_login_history_account_no` ON `tbl_login_history`(`account_no`);
CREATE INDEX `idx_feedback_account_no` ON `tbl_feedback`(`account_no`);
