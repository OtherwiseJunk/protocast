# Root orchestrator for the protocast monorepo.

.PHONY: check test dev install lex-check backend-check backend-test backend-install frontend-check frontend-install

check: lex-check backend-check frontend-check

test: backend-test

install: backend-install frontend-install

lex-check:
	cd backend && go run ./cmd/lexcheck

backend-check:
	$(MAKE) -C backend check

backend-test:
	$(MAKE) -C backend test

backend-install:
	$(MAKE) -C backend install

frontend-check:
	cd frontend && pnpm turbo run check

frontend-install:
	cd frontend && pnpm install

# Brings up backend + PDS stack for local dev.
dev:
	docker compose -f deploy/docker-compose.yml up -d
