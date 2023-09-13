.SILENT: ;

BIN=$(CURDIR)/bin
CMD=$(CURDIR)/cmd
GO=$(shell which go)
APP=Go-SSH-Server

keygen:
	$(GO) run $(CMD)/keygen/main.go

server:
	$(GO) run $(CMD)/server/main.go

exec:
	$(GO) run $(CMD)/client/main.go

shell:
	ssh -p 2023 -i $(CURDIR)/server_key.pem KMITL@localhost
