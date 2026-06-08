-- Migration 000001: Drop all tables (reverse of initial schema)
-- Order matters: drop tables without FK dependencies first

DROP TABLE IF EXISTS `tbl_transaction`;
DROP TABLE IF EXISTS `tbl_requests`;
DROP TABLE IF EXISTS `tbl_login_history`;
DROP TABLE IF EXISTS `tbl_feedback`;
DROP TABLE IF EXISTS `tbl_customer`;
DROP TABLE IF EXISTS `tbl_balance`;
DROP TABLE IF EXISTS `tbl_address`;
DROP TABLE IF EXISTS `tbl_account_type`;
DROP TABLE IF EXISTS `tbl_admin`;
DROP TABLE IF EXISTS `tbl_account`;
