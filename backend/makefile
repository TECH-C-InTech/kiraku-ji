# 必ず実行できるように
.PHONY: fmt lint test ci 

# formatを実行する
fmt: 
	go fmt ./...
	@echo "fmt complete"

# 静的解析を実行して、5分でタイムアウト
lint:
	golangci-lint run --timeout 5m

# 全パッケージのテストを実行する
test:
	go test ./...
	@echo "test complete"

# すべてのチェックをまとめて実行する
ci: fmt lint test
