ROOT_DIR:=../../
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all:
	./razee.sh

.PHONY: install_control
install_control:
	./razee.sh install_control

.PHONY: setup_user
setup_user:
	./razee.sh setup_user

.PHONY: setup_remotes
setup_remotes:
	./razee.sh setup_remotes

.PHONY: cleanup
cleanup:
	./razee.sh cleanup

.PHONY: cleanup_remotes
cleanup_remotes:
	./razee.sh cleanup_remotes
