#
# Copyright 2022 The Dapr Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

MODFILES := $(shell find . -name go.mod)
################################################################################
# Target: modtidy-all                                                          #
################################################################################

define modtidy-target
.PHONY: modtidy-$(1)
modtidy-$(1):
	cd $(shell dirname $(1)); go mod tidy -compat=1.20; cd -
endef

# Generate modtidy target action for each go.mod file
$(foreach MODFILE,$(MODFILES),$(eval $(call modtidy-target,$(MODFILE))))

# Enumerate all generated modtidy targets
TIDY_MODFILES:=$(foreach ITEM,$(MODFILES),modtidy-$(ITEM))

# Define modtidy-all action trigger to run make on all generated modtidy targets
.PHONY: modtidy-all
modtidy-all: $(TIDY_MODFILES)

# Replace all subdirectories with the desired dapr version.
# usage: DAPR_VERSION="github.com/$REPOSITORY/dapr $VERSION" make replace-all
################################################################################
# Target: replace-all                                                          #
################################################################################

define replace-target
.PHONY: replace-$(1)
replace-$(1):
	cd $(shell dirname $(1)); go mod edit -replace github.com/dapr/dapr=$(DAPR_VERSION); cd -
endef

# Generate replace target action for each go.mod file
$(foreach MODFILE,$(MODFILES),$(eval $(call replace-target,$(MODFILE))))

# Enumerate all generated replace targets
REPLACE_MODFILES:=$(foreach ITEM,$(MODFILES),replace-$(ITEM))

.PHONY: replace-all
replace-all: $(REPLACE_MODFILES)


# Executes a go get using the desired version.
# usage: DAPR_VERSION="github.com/$REPOSITORY/dapr@$HASH" make get-all
################################################################################
# Target: get-all                                                              #
################################################################################

define get-target
.PHONY: get-$(1)
get-$(1):
	cd $(shell dirname $(1)); go get $(DAPR_VERSION); cd -
endef

# Generate get target action for each go.mod file
$(foreach MODFILE,$(MODFILES),$(eval $(call get-target,$(MODFILE))))

# Enumerate all generated get targets
GET_MODFILES:=$(foreach ITEM,$(MODFILES),get-$(ITEM))

.PHONY: get-all
get-all: $(GET_MODFILES)

# Executes a go get using the latest dapr version (i.e github.com/dapr/dapr@master) and drop all replaces.
# usage: make update-latest
################################################################################
# Target: update-latest                                                              #
################################################################################

define latest-target
.PHONY: latest-$(1)
latest-$(1):
	cd $(shell dirname $(1));go mod edit -dropreplace github.com/dapr/dapr;go get github.com/dapr/dapr@master; go mod tidy; cd -
endef

# Generate latest target action for each go.mod file
$(foreach MODFILE,$(MODFILES),$(eval $(call latest-target,$(MODFILE))))

# Enumerate all generated mod targets
LATEST_MODFILES:=$(foreach ITEM,$(MODFILES),latest-$(ITEM))

.PHONY: update-latest
update-latest: $(LATEST_MODFILES)