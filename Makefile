SHELL := /usr/bin/env bash

lint:
	shellcheck -x bin/arch-wsl lib/bash/*.sh || true
	pwsh -NoLogo -NoProfile -Command "Install-Module PSScriptAnalyzer -Scope CurrentUser -Force; Invoke-ScriptAnalyzer -Path src/ps -Recurse -Severity Warning" || true

test:
	bats tests/bats || true

phase1:
	DEFAULT_USER?=erik DNS_MODE?=static sudo -E bin/arch-wsl phase1

phase2:
	DEFAULT_USER?=erik DNS_MODE?=static sudo -E bin/arch-wsl phase2
