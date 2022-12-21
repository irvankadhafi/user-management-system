SHELL:=/bin/bash


internal/model/mock/mock_user_repository.go:
	mockgen -destination=internal/model/mock/mock_user_repository.go -package=mock user-service/internal/model UserRepository
internal/model/mock/mock_rbac_repository.go:
	mockgen -destination=internal/model/mock/mock_rbac_repository.go -package=mock user-service/internal/model RBACRepository

mockgen: internal/model/mock/mock_rbac_repository.go

check-modd-exists:
	@modd --version > /dev/null

run: check-modd-exists
	@modd -f ./.modd/server.modd.conf