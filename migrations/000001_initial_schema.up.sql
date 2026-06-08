-- Migration 000001: Initial schema
-- Creates the full database schema based on the existing bank_db.sql

CREATE TABLE IF NOT EXISTS `tbl_account` (
  `account_no` INT(9) NOT NULL AUTO_INCREMENT,
  `username` CHAR(25) NOT NULL,
  `password` CHAR(25) NOT NULL,
  PRIMARY KEY (`account_no`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_account_type` (
  `account_no` INT(9) NOT NULL,
  `account_type` CHAR(10) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_address` (
  `account_no` INT(9) NOT NULL,
  `home_address` VARCHAR(100) NOT NULL,
  `city` CHAR(25) NOT NULL,
  `state` CHAR(25) NOT NULL,
  `pincode` INT(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_admin` (
  `admin_id` INT(11) NOT NULL AUTO_INCREMENT,
  `full_name` CHAR(25) NOT NULL,
  `mobile` CHAR(14) NOT NULL,
  `email` VARCHAR(50) NOT NULL,
  `password` CHAR(25) NOT NULL,
  PRIMARY KEY (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_balance` (
  `account_no` INT(9) NOT NULL,
  `account_type` VARCHAR(20) NOT NULL,
  `balance` INT(10) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_customer` (
  `account_no` INT(9) NOT NULL,
  `full_name` CHAR(100) NOT NULL,
  `gender` CHAR(10) NOT NULL,
  `birth_date` DATE NOT NULL,
  `mobile` CHAR(15) NOT NULL,
  `email` CHAR(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_feedback` (
  `feedback_id` INT(10) NOT NULL AUTO_INCREMENT,
  `account_no` INT(9) NOT NULL,
  `feedback` VARCHAR(1000) NOT NULL,
  `hearts` INT(6) DEFAULT NULL,
  `time` DATETIME NOT NULL,
  PRIMARY KEY (`feedback_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_login_history` (
  `token_id` INT(10) NOT NULL AUTO_INCREMENT,
  `account_no` INT(9) NOT NULL,
  `login_time` DATETIME NOT NULL,
  `logout_time` DATETIME DEFAULT NULL,
  PRIMARY KEY (`token_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_requests` (
  `request_id` INT(10) NOT NULL AUTO_INCREMENT,
  `account_no` INT(9) NOT NULL,
  `to_account` INT(9) NOT NULL,
  `amount` INT(10) NOT NULL,
  `message` VARCHAR(1000) NOT NULL,
  `hasViewed` TINYINT(1) NOT NULL DEFAULT 0,
  `status` CHAR(15) NOT NULL,
  `request_date` DATETIME NOT NULL,
  PRIMARY KEY (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `tbl_transaction` (
  `trans_id` INT(100) NOT NULL AUTO_INCREMENT,
  `trans_date` DATETIME NOT NULL,
  `amount` INT(100) NOT NULL,
  `trans_type` CHAR(10) NOT NULL,
  `purpose` VARCHAR(100) NOT NULL,
  `to_account` INT(9) NOT NULL,
  `account_no` INT(9) NOT NULL,
  `account_bal` INT(100) NOT NULL,
  PRIMARY KEY (`trans_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
