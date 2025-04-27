install-codegen:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate-dto: install-codegen
	oapi-codegen \
		-generate types \
		-package dto \
		-o internal/generated/dto.go \
		api/swagger.yaml
	go mod tidy

generate-mocks:
	mockgen -source=internal/storage/interfaces.go -destination=internal/generated/mocks/mock_interfaces.go -package=mocks