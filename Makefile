# Root orchestrator for the protocast monorepo.

.PHONY: check test dev lex-check backend-check backend-test admin-check protocast-check

check: lex-check backend-check admin-check protocast-check

test: backend-test

lex-check:
	cd backend && go run ./cmd/lexcheck

backend-check:
	$(MAKE) -C backend check

backend-test:
	$(MAKE) -C backend test

# TODO: wire up once admin/ exists
admin-check:
	@test -d admin && $(MAKE) -C admin check || echo "skip: admin/ not scaffolded yet"

# TODO: wire up once protocast/ exists
protocast-check:
	@test -d protocast && $(MAKE) -C protocast check || echo "skip: protocast/ not scaffolded yet"

# Brings up backend + PDS stack for local dev.
dev:
	docker compose -f deploy/docker-compose.yml up -d
