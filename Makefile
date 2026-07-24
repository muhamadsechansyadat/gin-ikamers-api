# ==============================================================================
# Load .env otomatis
# ==============================================================================
ifneq (,$(wildcard ./.env))
  include .env
  export
endif

# ==============================================================================
# PHONY - tandai target bukan file
# ==============================================================================
.PHONY: migrate-up migrate-down migrate-status migrate-reset migrate-redo migrate-create

# ==============================================================================
# Migration (goose)
# ==============================================================================
migrate-up:
	goose -dir=$(GOOSE_MIGRATION_DIR) up

migrate-down:
	goose -dir=$(GOOSE_MIGRATION_DIR) down

migrate-status:
	goose -dir=$(GOOSE_MIGRATION_DIR) status

migrate-reset:
	goose -dir=$(GOOSE_MIGRATION_DIR) reset

migrate-redo:
	goose -dir=$(GOOSE_MIGRATION_DIR) redo

migrate-create:
	goose -dir=$(GOOSE_MIGRATION_DIR) create $(name) sql