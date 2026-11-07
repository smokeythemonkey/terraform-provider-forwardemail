#!/usr/bin/env bash

# The import ID is "<domain>/<name>"; the alias name may itself contain
# slashes, which are kept intact.
terraform import forwardemail_alias.example example.com/postmaster
