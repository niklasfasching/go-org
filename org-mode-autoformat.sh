#!/usr/bin/env bash

set -e

fix=false

if [[ $1 == "--fix" ]]; then
	fix=true
	shift
fi

changes=false

for i in "$@"; do
	tmp=$(mktemp)
	go-org render "$i" org >"$tmp"

	if ! diff "$i" "$tmp"; then
		changes=true

		if [[ $fix == true ]]; then
			# overwrite input files
			echo "fixing..."
			mv "$tmp" "$i"
		else
			# show diff, do not overwrite
			# diff "$i" "$tmp"
			rm "$tmp"
		fi

	else
		rm "$tmp"
	fi

done

if [[ $changes == true ]]; then
	exit 1
else
	exit 0
fi
