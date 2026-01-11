-- +goose Up
-- +goose StatementBegin
INSERT INTO transfer_methods (name, display)
VALUES ('ewallet', 'E-Wallet');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM transfer_methods WHERE name = 'ewallet';
-- +goose StatementEnd
