.PHONY: vault-setup
vault-setup: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_kind.sh

.PHONY: vault-setup-multi
vault-setup-multi: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_kind.sh multi

.PHONY: vault-cleanup
vault-cleanup: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_kind.sh cleanup

.PHONY: configure-vault
configure-vault: vault-cleanup vault-setup

