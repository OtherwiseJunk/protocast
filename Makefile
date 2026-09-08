# Root orchestrator for the protocast monorepo.

.PHONY: check test dev lex-check backend-check backend-test frontend-check

check: lex-check backend-check frontend-check

test: backend-test

lex-check:
	cd backend && go run ./cmd/lexcheck

backend-check:
	$(MAKE) -C backend check

backend-test:
	$(MAKE) -C backend test

frontend-check:
	cd frontend && pnpm turbo run check

# Brings up backend + PDS stack for local dev.
dev:
	docker compose -f deploy/docker-compose.yml up -d
