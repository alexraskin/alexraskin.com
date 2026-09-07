#!/usr/bin/env bash
# Strip a photo's metadata, upload it to R2, and print the JSON entry to paste
# into data/franzbroetchen.json.
#
# One object per photo. Cloudflare Image Transformations does the resizing and
# the format negotiation at request time, off the /cdn-cgi/image/ prefix on the
# same domain, so there are no variants to generate, commit, or keep in sync.
#
# The key carries the content hash, so the object is immutable and can be served
# with a year-long max-age. Re-uploading the same photo yields the same key;
# a different photo yields a new one rather than overwriting a live URL.

set -euo pipefail

readonly BUCKET="${R2_BUCKET:-cdn-alexraskin}"
readonly PREFIX="franzbroetchen"
# The page never renders wider than 1320 CSS pixels; this leaves room for a 2x
# display without storing a 12-megapixel phone photo.
readonly MAX_WIDTH=2640
readonly QUALITY=90

DRY_RUN=0
# Set by main, removed on exit. Global because the EXIT trap outlives main's
# locals, and under `set -u` an unset name there aborts the script.
work=""

usage() {
	cat >&2 <<-EOF
		usage: mise run add-review [--dry-run] <photo> [stem]

		  photo      the source image, at any size
		  stem       basename for the uploaded key; defaults to a slug of the photo
		  --dry-run  strip and hash, print the entry, upload nothing

		example: mise run add-review ~/Downloads/IMG_5044.jpg elbgold-eppendorf

		A review can show more than one photo: run this once per photo with
		stems that differ, e.g. elbgold-eppendorf-1 and elbgold-eppendorf-2,
		and list both objects under "photos".

		Uploads go to s3://$BUCKET/$PREFIX/ and are served from
		\$CDN_BASE_URL (default https://cdn.alexraskin.com).

		Credentials are read the way the AWS CLI normally reads them. R2 needs:

		  R2_ACCOUNT_ID          your Cloudflare account id
		  AWS_PROFILE            a profile holding the R2 access key pair, or
		  AWS_ACCESS_KEY_ID      the R2 token's access key id, and
		  AWS_SECRET_ACCESS_KEY  its secret
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
# Cloudflare would also drop most of this on delivery — its metadata parameter
# defaults to "copyright", which discards GPS. That default is a delivery-time
# setting on someone else's product, though, and it is one dashboard toggle
# (metadata=keep, or flexible variants letting a caller ask for it) away from
# serving the location back. Stripping before the upload means the bucket never
# holds the coordinates in the first place, so no delivery setting can leak them.
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
	local -a args=()
	local arg
	for arg in "$@"; do
		case "$arg" in
		--dry-run) DRY_RUN=1 ;;
		-h | --help) usage ;;
		-*)
			echo "add-review: unknown flag: $arg" >&2
			usage
			;;
		*) args+=("$arg") ;;
		esac
	done
	set -- "${args[@]+"${args[@]}"}"

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

	if ((!DRY_RUN)); then
		command -v aws >/dev/null 2>&1 || {
			echo "add-review: needs the AWS CLI to reach R2 (brew install awscli / apt install awscli)" >&2
			echo "            or re-run with --dry-run to strip without uploading" >&2
			exit 69
		}
		[[ -n ${R2_ACCOUNT_ID:-} ]] || {
			echo "add-review: R2_ACCOUNT_ID is not set, see --help" >&2
			exit 78
		}
	fi

	work=$(mktemp -d)
	local clean="$work/$stem.jpg"

	# ">" only shrinks: a photo already narrower than MAX_WIDTH is left alone
	# rather than upscaled.
	im "${photo}[0]" -auto-orient -resize "${MAX_WIDTH}x>" -strip +profile '*' \
		-quality "$QUALITY" "$clean"

	assert_clean "$clean"

	local size
	size=$(im_identify -format '%wx%h' "$clean")

	local key="$PREFIX/$stem.$(sha256sum "$clean" | cut -c1-12).jpg"

	printf '  %s  %s  %s\n' "$size" "$(du -h "$clean" | cut -f1)" "$key"

	if ((DRY_RUN)); then
		printf '  would upload  %s\n' "s3://$BUCKET/$key"
	else
		aws --endpoint-url "https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com" \
			s3api put-object \
			--bucket "$BUCKET" \
			--key "$key" \
			--body "$clean" \
			--content-type image/jpeg \
			--cache-control 'public, max-age=31536000, immutable' \
			>/dev/null
		printf '  uploaded      %s\n' "s3://$BUCKET/$key"
	fi

	local entry
	entry=$(jq -n \
		--arg key "$key" \
		--argjson width "${size%x*}" \
		--argjson height "${size#*x}" \
		--arg date "$(date +%F)" \
		'{
			place: "",
			location: "",
			date: $date,
			rating: 0,
			photos: [{key: $key, width: $width, height: $height}],
			note: ""
		}')

	printf '\nadd to data/franzbroetchen.json:\n\n%s\n' "$entry" | sed -e '3,$s/^/  /'

	if ((DRY_RUN)); then
		echo
		echo "add-review: dry run, nothing was uploaded" >&2
	fi
}

trap '[[ -n $work ]] && rm -rf "$work"' EXIT

main "$@"
