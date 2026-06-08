-- Migration 000003: Add foreign key constraints referencing tbl_account(account_no)
-- ON DELETE RESTRICT prevents accidental deletion of accounts with related data
-- ON UPDATE CASCADE propagates account_no changes

ALTER TABLE `tbl_account_type`
  ADD CONSTRAINT `fk_account_type_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_address`
  ADD CONSTRAINT `fk_address_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_balance`
  ADD CONSTRAINT `fk_balance_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_customer`
  ADD CONSTRAINT `fk_customer_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_feedback`
  ADD CONSTRAINT `fk_feedback_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_login_history`
  ADD CONSTRAINT `fk_login_history_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_requests`
  ADD CONSTRAINT `fk_requests_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `tbl_transaction`
  ADD CONSTRAINT `fk_transaction_account`
  FOREIGN KEY (`account_no`) REFERENCES `tbl_account`(`account_no`)
  ON DELETE RESTRICT ON UPDATE CASCADE;
