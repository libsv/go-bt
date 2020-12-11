## Default repository domain name
ifndef GIT_DOMAIN
	override GIT_DOMAIN=github.com
endif

## Set if defined (alias variable for ease of use)
ifdef branch
	override REPO_BRANCH=$(branch)
	export REPO_BRANCH
endif

## Do we have git available?
HAS_GIT := $(shell command -v git 2> /dev/null)

ifdef HAS_GIT
	## Do we have a repo?
	HAS_REPO := $(shell git rev-parse --is-inside-work-tree 2> /dev/null)
	ifdef HAS_REPO
		## Automatically detect the repo owner and repo name (for local use with Git)
		REPO_NAME=$(shell basename "$(shell git rev-parse --show-toplevel 2> /dev/null)")
		OWNER=$(shell git config --get remote.origin.url | sed 's/git@$(GIT_DOMAIN)://g' | sed 's/\/$(REPO_NAME).git//g')
		REPO_OWNER=$(shell echo $(OWNER) | tr A-Z a-z)
		VERSION_SHORT=$(shell git describe --tags --always --abbrev=0)
		export REPO_NAME, REPO_OWNER, VERSION_SHORT
	endif
endif

## Set the distribution folder
ifndef DISTRIBUTIONS_DIR
	override DISTRIBUTIONS_DIR=./dist
endif
export DISTRIBUTIONS_DIR

help: ## Show this help message
	@egrep -h '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

release:: ## Full production release (creates release in Github)
	@test $(github_token)
	@export GITHUB_TOKEN=$(github_token) && goreleaser --rm-dist

release-test: ## Full production test release (everything except deploy)
	@goreleaser --skip-publish --rm-dist

release-snap: ## Test the full release (build binaries)
	@goreleaser --snapshot --skip-publish --rm-dist

replace-version: ## Replaces the version in HTML/JS (pre-deploy)
	@test $(version)
	@test "$(path)"
	@find $(path) -name "*.html" -type f -exec sed -i '' -e "s/{{version}}/$(version)/g" {} \;
	@find $(path) -name "*.js" -type f -exec sed -i '' -e "s/{{version}}/$(version)/g" {} \;

tag: ## Generate a new tag and push (tag version=0.0.0)
	@test $(version)
	@git tag -a v$(version) -m "Pending full release..."
	@git push origin v$(version)
	@git fetch --tags -f

tag-remove: ## Remove a tag if found (tag-remove version=0.0.0)
	@test $(version)
	@git tag -d v$(version)
	@git push --delete origin v$(version)
	@git fetch --tags

tag-update: ## Update an existing tag to current commit (tag-update version=0.0.0)
	@test $(version)
	@git push --force origin HEAD:refs/tags/v$(version)
	@git fetch --tags -f

update-releaser:  ## Update the goreleaser application
	@brew update
	@brew upgrade goreleaser
