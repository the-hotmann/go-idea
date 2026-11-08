AMD64        := v1 v2 v3 v4
COUNT        := 3
BENCH        := .
GOEXPERIMENT := simd
BENCHSTAT    := $(or $(shell command -v benchstat),$(shell go env GOPATH)/bin/benchstat)

export GOEXPERIMENT

.PHONY: bench
bench:
	@out=$$(mktemp -d) && \
	for v in $(AMD64); do \
		echo "GOAMD64=$$v GOEXPERIMENT=$$GOEXPERIMENT: go test -bench='$(BENCH)' -benchmem -count=$(COUNT)"; \
		GOAMD64=$$v go test -run='^$$' -bench='$(BENCH)' -benchmem -count=$(COUNT) . > "$$out/$$v" || { cat "$$out/$$v"; exit 1; }; \
	done && \
	cd "$$out" && $(BENCHSTAT) $(AMD64) && \
	echo "raw results: $$out"
