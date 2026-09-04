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

		A review can show more than one photo: run this once per photo with
		stems that differ, e.g. elbgold-eppendorf-1 and elbgold-eppendorf-2,
		and list both paths under "photos".
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

# Phone photos carry EXIF the page has no use for and the internet has no
# business with: the camera, the timestamp, and on iPhones a GPS fix accurate to
# a few metres. -auto-orient bakes the rotation into the pixels first, so
# dropping the metadata cannot leave the photo on its side; -strip removes EXIF,
# IPTC and the embedded thumbnail, and +profile '*' takes the colour and XMP
# profiles with it.
#
# assert_clean is the belt to that braces: an ImageMagick build that quietly
# kept something, or a future edit that drops a flag, should fail the run rather
# than publish a home address.
assert_clean() {
	local file="$1"

	if im_identify -verbose "$file" | grep -q '^ *exif:'; then
		echo "add-review: $file still has EXIF, refusing to publish it" >&2
		exit 65
	fi

	# identify only reports what it can parse, so also look at the bytes for the
	# markers a stripped file has no reason to contain.
	if LC_ALL=C grep -qaE 'GPSLatitude|DateTimeOriginal|Exif' "$file"; then
		echo "add-review: $file still has metadata markers, refusing to publish it" >&2
		exit 65
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
		im "${photo}[0]" -auto-orient -resize "${width}x" -strip +profile '*' \
			-quality "$AVIF_QUALITY" "$OUT_DIR/$stem-$width.avif"
		im "${photo}[0]" -auto-orient -resize "${width}x" -strip +profile '*' \
			-interlace Plane -quality "$JPEG_QUALITY" "$OUT_DIR/$stem-$width.jpg"

		assert_clean "$OUT_DIR/$stem-$width.avif"
		assert_clean "$OUT_DIR/$stem-$width.jpg"

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
		    "photos": ["/$OUT_DIR/$stem.jpg"],
		    "note": ""
		  }
	EOF
}

main "$@"
