install-codegen:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate-dto: install-codegen
	oapi-codegen \
		-generate types \
		-package dto \
		-o internal/generated/dto.go \
		api/swagger.yaml
	go mod tidy