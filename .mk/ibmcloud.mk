IBMCLOUD:=/usr/local/bin/ibmcloud

.PHONY: ibmcloud-login
ibmcloud-login: $(IBMCLOUD)
	$(IBMCLOUD) login --apikey ${DOCKER_PASSWORD}

.PHONY: ibmcloud-image-list
ibmcloud-image-list: $(IBMCLOUD) ibmcloud-login
	$(IBMCLOUD) cr image-list --restrict ${DOCKER_NAMESPACE}

.PHONY: ibmcloud-image-rm
ibmcloud-image-rm: $(IBMCLOUD) ibmcloud-login
	$(IBMCLOUD) cr image-rm ${IMG} || true
