.PHONY: all
all: check checkscripts test

.PHONY: test
test:
	@echo "==> Testing all packages with coverage"
	@rm -rf .results
	@mkdir .results
	gotestsum \
		--junitfile .results/results.xml \
		--jsonfile .results/results.json \
		--format testname \
		-- -coverprofile=.results/cover.out ./...
	@echo "==> Checking Coverage"
	@go-test-coverage --config=./.testcoverage.yml

.PHONY: tidy
tidy:
	@echo "==> Tidy main module"
	@go mod tidy

.PHONY: checkscripts
checkscripts:
	@echo "==> Linting scripts..."
	@find . -type f -name '*.sh' | xargs shellcheck

.PHONY: check
check:
	@echo "==> Linting source code..."
	@golangci-lint run

	@echo "==> Checking Go mod.."
	@GO111MODULE=on $(MAKE) -s tidy
	@if (git status --porcelain | grep -Eq "go\.(mod|sum)"); then \
		echo go.mod or go.sum needs updating; \
		git --no-pager diff go.mod; \
		git --no-pager diff go.sum; \
		exit 1; fi

