.PHONY: release push help

HELM_CHART := helm/mortar/Chart.yaml
HELM_VALUES := helm/mortar/values.yaml

help:
	@echo "Usage:"
	@echo "  make release TYPE=<major|minor|patch>  # Create a new release"
	@echo "  make push                               # Push backend and portal tags"
	@echo ""
	@echo "Examples:"
	@echo "  make release TYPE=major   # 1.2.3 -> 2.0.0"
	@echo "  make release TYPE=minor   # 1.2.3 -> 1.3.0"
	@echo "  make release TYPE=patch   # 1.2.3 -> 1.2.4"
	@echo "  make push                 # Push tags to trigger CI/CD"

release:
	@if [ -z "$(TYPE)" ]; then \
		echo "Error: TYPE not specified. Use: make release TYPE=<major|minor|patch>"; \
		exit 1; \
	fi
	@if [ "$(TYPE)" != "major" ] && [ "$(TYPE)" != "minor" ] && [ "$(TYPE)" != "patch" ]; then \
		echo "Error: TYPE must be one of: major, minor, patch"; \
		exit 1; \
	fi
	@echo "==> Getting latest version from git tags..."
	@LATEST_TAG=$$(git tag --list 'v*' | sed 's/^v//' | sort -V | tail -1); \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "No existing tags found, starting from 0.0.0"; \
		LATEST_TAG="0.0.0"; \
	else \
		echo "Latest tag: $$LATEST_TAG"; \
	fi; \
	MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1); \
	MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
	PATCH=$$(echo $$LATEST_TAG | cut -d. -f3); \
	if [ "$(TYPE)" = "major" ]; then \
		MAJOR=$$((MAJOR + 1)); \
		MINOR=0; \
		PATCH=0; \
	elif [ "$(TYPE)" = "minor" ]; then \
		MINOR=$$((MINOR + 1)); \
		PATCH=0; \
	else \
		PATCH=$$((PATCH + 1)); \
	fi; \
	NEW_VERSION="$$MAJOR.$$MINOR.$$PATCH"; \
	echo "==> New version: $$NEW_VERSION"; \
	echo ""; \
	echo "==> Updating Helm Chart.yaml..."; \
	sed -i.bak "s/^version: .*/version: $$NEW_VERSION/" $(HELM_CHART); \
	sed -i.bak "s/^appVersion: .*/appVersion: \"$$NEW_VERSION\"/" $(HELM_CHART); \
	rm -f $(HELM_CHART).bak; \
	echo "==> Updating Helm values.yaml Docker image tags and crossplane package..."; \
	awk -v ver="v$$NEW_VERSION" ' \
		/^backend:/ { in_backend=1; in_frontend=0; in_crossplane=0 } \
		/^frontend:/ { in_frontend=1; in_backend=0; in_crossplane=0 } \
		/^crossplane:/ { in_crossplane=1; in_backend=0; in_frontend=0 } \
		/^[a-zA-Z]/ && !/^backend:/ && !/^frontend:/ && !/^crossplane:/ { in_backend=0; in_frontend=0; in_crossplane=0 } \
		/^    image:/ && (in_backend || in_frontend) { in_image=1 } \
		/^    package:/ && in_crossplane { in_package=1 } \
		/^    [a-zA-Z]/ && !/^    image:/ && !/^    package:/ { in_image=0; in_package=0 } \
		/^        tag:/ && (in_image || in_package) { print "        tag: \"" ver "\""; next } \
		{ print } \
	' $(HELM_VALUES) > $(HELM_VALUES).tmp; \
	mv $(HELM_VALUES).tmp $(HELM_VALUES); \
	echo "==> Creating git tag v$$NEW_VERSION in portal submodule..."; \
	git -C portal tag -a "v$$NEW_VERSION" -m "Release version $$NEW_VERSION"; \
	echo "==> Creating git tag v$$NEW_VERSION in crossplane submodule..."; \
	git -C crossplane tag -a "v$$NEW_VERSION" -m "Release version $$NEW_VERSION"; \
	echo "==> Creating git tag v$$NEW_VERSION in backend..."; \
	git add $(HELM_CHART) $(HELM_VALUES); \
	git commit -m "chore: bump version to $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release version $$NEW_VERSION"; \
	echo ""; \
	echo "✓ Release v$$NEW_VERSION created successfully!"; \
	echo ""; \
	echo "Next steps:"; \
	echo "  make push"

push:
	@echo "==> Pushing backend repository..."
	@git push origin main
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
	if [ -n "$$LATEST_TAG" ]; then \
		echo "==> Pushing backend tag $$LATEST_TAG..."; \
		git push origin $$LATEST_TAG; \
		echo "==> Pushing portal tag $$LATEST_TAG..."; \
		git -C portal push origin $$LATEST_TAG; \
		echo "==> Pushing crossplane tag $$LATEST_TAG..."; \
		git -C crossplane push origin $$LATEST_TAG; \
		echo ""; \
		echo "✓ Tags pushed successfully!"; \
		echo "✓ CI/CD pipelines will now build and publish all components"; \
	else \
		echo "Error: No tags found. Run 'make release' first."; \
		exit 1; \
	fi
