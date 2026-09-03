#!/usr/bin/env bash
# Turn one photo into the AVIF/JPEG variants the Franzbrötchen page serves, and
# print the JSON entry to paste into data/franzbroetchen.json.
#
# Only the derivatives are committed; keep the original wherever your photos
# already live. Nothing reads the widths from a config: the page discovers
# whatever variants exist on disk, so adding a size later is just another run.

set -euo pipefail

readonly OUT_DIR="assets/images/franzbroetchen"
readonly WIDTHS=(1320 672)
readonly AVIF_QUALITY=55
readonly JPEG_QUALITY=80

usage() {
	cat >&2 <<-EOF
		usage: mise run add-review <photo> [stem]

		  photo  the source image, at any size
		  stem   basename for the generated files; defaults to a slug of the photo

		example: mise run add-review ~/Downloads/IMG_5044.jpg elbgold-eppendorf
	EOF
	exit 64
}

# ImageMagick 7 ships "magick"; the 6 still on some machines only has "convert".
im() {
	if command -v magick >/dev/null 2>&1; then
		magick "$@"
	else
		convert "$@"
	fi
}

im_identify() {
	if command -v magick >/dev/null 2>&1; then
		magick identify "$@"
	else
		identify "$@"
	fi
}

slugify() {
	basename "$1" |
		sed -e 's/\.[^.]*$//' |
		tr '[:upper:] ._' '[:lower:]---' |
		sed -e 's/[^a-z0-9-]//g' -e 's/-\{2,\}/-/g' -e 's/^-//' -e 's/-$//'
}

main() {
	[[ $# -ge 1 && $# -le 2 ]] || usage

	local photo="$1"
	local stem="${2:-$(slugify "$1")}"

	if ! command -v magick >/dev/null 2>&1 && ! command -v convert >/dev/null 2>&1; then
		echo "add-review: needs ImageMagick (brew install imagemagick / apt install imagemagick)" >&2
		exit 69
	fi

	[[ -f $photo ]] || {
		echo "add-review: no such file: $photo" >&2
		exit 66
	}
	[[ -n $stem ]] || {
		echo "add-review: could not derive a name from $photo, pass one explicitly" >&2
		exit 64
	}

	local source_width
	source_width=$(im_identify -format '%w' "${photo}[0]")

	# Upscaling only inflates the download, so a width past the source is
	# skipped; a photo smaller than every width still gets encoded once.
	local -a widths=()
	local width
	for width in "${WIDTHS[@]}"; do
		((width <= source_width)) && widths+=("$width")
	done
	((${#widths[@]})) || widths=("$source_width")

	mkdir -p "$OUT_DIR"

	for width in "${widths[@]}"; do
		im "${photo}[0]" -auto-orient -resize "${width}x" -strip \
			-quality "$AVIF_QUALITY" "$OUT_DIR/$stem-$width.avif"
		im "${photo}[0]" -auto-orient -resize "${width}x" -strip \
			-interlace Plane -quality "$JPEG_QUALITY" "$OUT_DIR/$stem-$width.jpg"

		printf '  %-52s %6s\n' \
			"$OUT_DIR/$stem-$width.avif" "$(du -h "$OUT_DIR/$stem-$width.avif" | cut -f1)"
		printf '  %-52s %6s\n' \
			"$OUT_DIR/$stem-$width.jpg" "$(du -h "$OUT_DIR/$stem-$width.jpg" | cut -f1)"
	done

	cat <<-EOF

		add to data/franzbroetchen.json:

		  {
		    "place": "",
		    "location": "",
		    "date": "$(date +%F)",
		    "rating": 0,
		    "photo": "/$OUT_DIR/$stem.jpg",
		    "note": ""
		  }
	EOF
}

main "$@"
