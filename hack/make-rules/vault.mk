.PHONY: vault-setup
vault-setup: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_single_cluster.sh

.PHONY: vault-setup-kind-multi
vault-setup-kind-multi: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_kind_multi.sh

.PHONY: vault-cleanup
vault-cleanup: $(TOOLBIN)/vault $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./configure_vault_single_cluster.sh cleanup

.PHONY: configure-vault
configure-vault: vault-cleanup vault-setup

