-- +goose Up
INSERT INTO transfer_methods
(name, display)
VALUES 
('GROUP_EXPENSE', 'Group Expense'),
('bank', 'Bank'),
('cash', 'Cash'),
('app', 'App Transfer');


-- +goose Down
-- DELETE FROM transfer_methods; -- Warning: this deletes all rows
