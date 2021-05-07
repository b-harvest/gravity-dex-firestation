#!/usr/bin/make -f

###############################################################################
###                                Build                                    ###
###############################################################################

install: go.sum
	@echo "Installing tester binary..."
	@go install ./cmd/tester

.PHONY: install

###############################################################################
###                               Localnet                                  ###
###############################################################################

localnet: 
	@echo "Bootstraping a single local testnet..."
	./scripts/localnet.sh

.PHONY: localnet