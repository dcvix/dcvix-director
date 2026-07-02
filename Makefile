MAIN_NAME=dcvix-director

DIST_DIR=dist

# Version information
VERSION?=$(shell (git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0") | sed 's/^v//')
RELEASE=1
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build variables
BINARY_NAME=$(MAIN_NAME)
GO=$(shell which go)
LDFLAGS="-X github.com/dcvix/$(MAIN_NAME)/internal/version.Version=$(VERSION) \
         -X github.com/dcvix/$(MAIN_NAME)/internal/version.Commit=$(COMMIT) \
         -X github.com/dcvix/$(MAIN_NAME)/internal/version.BuildTime=$(BUILD_TIME)"

# Platform-specific variables
LINUX_AMD64_BINARY=$(MAIN_NAME)
LINUX_AMD64_DIR=$(MAIN_NAME)-v$(VERSION)-linux-amd64

# RPM variables
RPM_NAME=$(BINARY_NAME)
RPM_VERSION=$(VERSION)
RPM_RELEASE=$(RELEASE)
RPM_TOPDIR=$(CURDIR)/rpmbuild
RPM_SOURCES=$(RPM_TOPDIR)/SOURCES
RPM_SPECS=$(RPM_TOPDIR)/SPECS

# Debian package variables
DEB_NAME=$(BINARY_NAME)
DEB_VERSION=$(VERSION)
DEB_RELEASE=$(RELEASE)
DEB_TOPDIR=$(CURDIR)/debbuild
DEB_SOURCE=$(DEB_TOPDIR)/$(DEB_NAME)-$(DEB_VERSION)
DEB_CONTROL=$(DEB_SOURCE)/DEBIAN
DEB_BINARY=$(DEB_SOURCE)/usr/bin
DEB_CONFIG=$(DEB_SOURCE)/etc/$(DEB_NAME)
DEB_SYSTEMD=$(DEB_SOURCE)/etc/systemd/system
DEB_LOG=$(DEB_SOURCE)/var/log/$(DEB_NAME)
DEB_DOC=$(DEB_SOURCE)/usr/share/doc/$(DEB_NAME)

# Frontend variables
NPM=npm
FRONTEND_DIR=frontend

# Build application and frontend
.PHONY: build
build: frontend bin
	cp README.md LICENSE.md $(DIST_DIR)/$(LINUX_AMD64_DIR)/
	cd $(DIST_DIR) && tar czf $(LINUX_AMD64_DIR).tar.gz $(LINUX_AMD64_DIR)

.PHONY: bin
bin: FORCE
	CGO_ENABLED=1 GOOS=linux $(GO) build \
		-trimpath -ldflags $(LDFLAGS) \
		-o $(DIST_DIR)/$(LINUX_AMD64_DIR)/$(LINUX_AMD64_BINARY) \
		./cmd/$(MAIN_NAME)
# Force target to make sure it is always rebuilt
FORCE: ;

## audit: run quality control checks
.PHONY: audit
audit:
	$(GO) mod tidy -diff
	$(GO) mod verify
	test -z "$(shell gofmt -l .)"
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest -checks=all ./...
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Show version
.PHONY: version
version:
	@echo $(VERSION)

# Frontend targets
.PHONY: frontend-deps
frontend-deps:
	cd $(FRONTEND_DIR) && $(NPM) install

.PHONY: frontend-dev
frontend-dev: frontend-deps
	cd $(FRONTEND_DIR) && $(NPM) run dev

.PHONY: frontend
frontend: frontend-deps
	cd $(FRONTEND_DIR) && $(NPM) run build

# Clean build artifacts
.PHONY: clean
clean: rpm-clean deb-clean
	rm -rf $(DIST_DIR)/
	cd $(FRONTEND_DIR) && rm -rf node_modules/ $(DIST_DIR)/

# Run the application
.PHONY: run
run: build
	$(DIST_DIR)/$(LINUX_AMD64_DIR)/$(LINUX_AMD64_BINARY)

# Run tests
.PHONY: test
test:
	$(GO) test ./... ;

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod tidy ;
	$(GO) mod verify ;


################# Manual INSTALLATION ################

# Install the application
.PHONY: install
install: build
	install -m 755 $(DIST_DIR)/$(LINUX_AMD64_DIR)/$(LINUX_AMD64_BINARY) /usr/bin/$(BINARY_NAME)
	install -m 644 internal/config/dcvix-director.conf.default /etc/dcvix-director.conf
	# Change log directory to /var/log/dcvix-agent
	sed -i 's|directory = log|directory = /var/log/dcvix-director|' /etc/dcvix-director.conf
	install -m 644 contrib/systemd/dcvix-director.service /lib/systemd/system/
	mkdir -p /var/log/dcvix-director
	systemctl daemon-reload
	@echo "Installation complete. Edit /etc/dcvix-director.conf and start the service with:"
	@echo "systemctl start dcvix-director"

.PHONY: uninstall
uninstall:
	systemctl stop dcvix-director
	systemctl disable dcvix-director
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f /etc/dcvix-director.conf
	rm -f /lib/systemd/system/dcvix-director.service
	rm -f /var/log/dcvix-director
	systemctl daemon-reload

################# RPM PACKAGE ################

# Build RPM package
.PHONY: rpm
rpm: rpm-prep
	cp contrib/rpm/$(RPM_NAME).spec $(RPM_SPECS)/
	rpmbuild --define "_topdir $(RPM_TOPDIR)" --define "pkg_version $(RPM_VERSION)" -ba $(RPM_SPECS)/$(RPM_NAME).spec
	mkdir -p $(DIST_DIR)
	cp $(RPM_TOPDIR)/RPMS/*/*.rpm $(DIST_DIR)/
	cp $(RPM_TOPDIR)/SRPMS/*.rpm $(DIST_DIR)/

# Prepare source for RPM
.PHONY: rpm-prep
rpm-prep:
	mkdir -p $(RPM_SOURCES)
	mkdir -p $(RPM_SPECS)
	mkdir -p $(RPM_SOURCES)/frontend
	cp -r README.md LICENSE.md Makefile cmd/ contrib/ internal/ go.mod go.sum $(RPM_SOURCES)
	cp -r  frontend $(RPM_SOURCES)/
	# Change log directory to /var/log/dcvix-director
	sed 's|directory = log|directory = /var/log/dcvix-director|' internal/config/dcvix-director.conf.default > $(RPM_SOURCES)/dcvix-director.conf
	tar --transform 's,^,$(RPM_NAME)-$(RPM_VERSION)/,' -C $(RPM_SOURCES) -czf \
		$(RPM_SOURCES)/$(RPM_NAME)-$(RPM_VERSION).tar.gz \
		README.md LICENSE.md Makefile dcvix-director.conf cmd/ contrib/ frontend/ internal/ go.mod go.sum

# Clean RPM build artifacts
.PHONY: rpm-clean
rpm-clean:
	rm -rf $(RPM_TOPDIR)

################# DEB PACKAGE ################

# Build Debian package
.PHONY: deb
deb: deb-prep
	cd $(DEB_TOPDIR) && dpkg-deb --build $(DEB_NAME)-$(DEB_VERSION)
	mkdir -p $(DIST_DIR)
	mv $(DEB_TOPDIR)/$(DEB_NAME)-$(DEB_VERSION).deb $(DIST_DIR)/

# Prepare source for Debian package
.PHONY: deb-prep
deb-prep: build
	mkdir -p $(DEB_CONTROL)
	mkdir -p $(DEB_BINARY)
	mkdir -p $(DEB_CONFIG)
	mkdir -p $(DEB_SYSTEMD)
	mkdir -p $(DEB_LOG)
	mkdir -p $(DEB_DOC)
	# Copy binary
	cp $(DIST_DIR)/$(LINUX_AMD64_DIR)/$(LINUX_AMD64_BINARY) $(DEB_BINARY)/
	# Copy config and change log directory to /var/log/dcvix-director
	sed 's|directory = log|directory = /var/log/dcvix-director|' internal/config/dcvix-director.conf.default > $(DEB_CONFIG)/dcvix-director.conf
	# Copy systemd service
	cp contrib/systemd/$(BINARY_NAME).service $(DEB_SYSTEMD)/
	# Copy documentation
	cp README.md $(DEB_DOC)/
	cp LICENSE.md $(DEB_DOC)/
	# Process and copy Debian control files
	sed -e 's/@PACKAGE@/$(DEB_NAME)/g' -e 's/@VERSION@/$(DEB_VERSION)-$(DEB_RELEASE)/g' \
		contrib/deb/control.in > $(DEB_CONTROL)/control
	cp contrib/deb/copyright $(DEB_CONTROL)/
	cp contrib/deb/postinst $(DEB_CONTROL)/
	cp contrib/deb/prerm $(DEB_CONTROL)/
	cp contrib/deb/postrm $(DEB_CONTROL)/
	# Make scripts executable
	chmod 755 $(DEB_CONTROL)/postinst
	chmod 755 $(DEB_CONTROL)/prerm
	chmod 755 $(DEB_CONTROL)/postrm

# Clean Debian build artifacts
.PHONY: deb-clean
deb-clean:
	rm -rf $(DEB_TOPDIR)
