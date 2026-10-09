.PHONY: dev dev-server dev-web build build-server build-web clean

dev:
	@$(MAKE) -j2 dev-server dev-web

dev-server:
	cd server && go run .

dev-web:
	cd web && npm run dev

build: build-web build-server

build-server:
	cd server && go build -o email-app .

build-web:
	cd web && npm run build

clean:
	rm -f server/email-app
	rm -rf web/dist
