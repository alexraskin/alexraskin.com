# alexraskin.com

Source for my personal website.

```bash
mise run dev      # local, live template reload
mise run docker-up # docker
```

## Deploying

Every push to `main` builds `ghcr.io/alexraskin/alexraskin.com:main-<epoch>-<sha>`,
which Flux rolls out to staging — reachable on the tailnet only, at
`alexraskin-staging`. Tagging `vX.Y.Z` builds `latest`, the tag and the commit
SHA, and that is what production tracks. The two tag shapes never overlap, so
staging can never promote itself.

## Adding a Franzbrötchen review

Review photos are not in this repository. `add-review` strips the metadata,
uploads one object per photo to the `cdn-alexraskin` R2 bucket, and prints the
JSON entry. Nothing is written into the working tree, and there are no size
variants to generate — Cloudflare resizes on request.

```sh
export R2_ACCOUNT_ID=<cloudflare account id>
export AWS_PROFILE=r2          # or AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY

mise run add-review ~/Downloads/IMG_5056.JPG mutterland-hamburg-1
mise run add-review ~/Downloads/IMG_5057.JPG mutterland-hamburg-2
```

Pass `--dry-run` to strip and print the entry without uploading.

Then append the printed object to `data/franzbroetchen.json` and fill in the
empty fields:

```json
{
  "place": "Bäckerei Name",
  "location": "Hamburg, Germany",
  "date": "2025-09-01",
  "rating": 4,
  "photos": [
    { "key": "franzbroetchen/photo.<hash>.jpg", "width": 2640, "height": 1980 }
  ],
  "note": "Optional.",
  "url": "https://optional-bakery-link"
}
```

`width` and `height` are the stored object's own size. The page derives the
rendered box from them so the layout does not shift while the photo loads, and
never asks for a width the object cannot fill.

### Photo metadata

Phone photos carry a GPS fix accurate to a few metres. `add-review` strips EXIF,
IPTC and the embedded thumbnail, then re-checks the result and **refuses to
upload** if any marker survived.

Cloudflare would drop most of it on delivery anyway — the `metadata` parameter
defaults to `copyright`, which discards GPS. That is a delivery-time setting on
someone else's product though, one toggle away from `metadata=keep`. Stripping
before upload means the bucket never holds the coordinates, so no delivery
setting can leak them.

## Serving

Photos are stored once in `cdn-alexraskin` and resized at request time by
[Cloudflare Image Transformations](https://developers.cloudflare.com/images/transform-images/),
off the `/cdn-cgi/image/` prefix on `cdn.alexraskin.com`:

```
https://cdn.alexraskin.com/cdn-cgi/image/width=672,format=auto,quality=82/franzbroetchen/photo.<hash>.jpg
```

`format=auto` negotiates AVIF or WebP per request, so the page ships a plain
`<img>` with a srcset rather than a `<picture>` with hand-encoded sources.

Requirements: Image Transformations enabled for the `alexraskin.com` zone, and
`cdn.alexraskin.com` attached to the bucket as an R2 custom domain (the source
image must sit on the same zone that serves the transformation).

Billing is per unique transformation. Two widths per photo, and `format=auto`
counts once no matter how many formats are served — so the whole page is two
transformations per photo against the Images Free plan's 5,000 a month.

The photo CDN base is configured in `main.go`. Photo keys carry a content hash.

Favicons, fonts, CSS and the Open Graph card stay embedded in the binary — the
site still renders without the bucket, only the review photos need it.

## Request handling

The server uses Go's standard HTTP router and file server. Assets are cached for
one week; HTML uses `Cache-Control: no-cache`. Rate limiting and visitor-IP
handling belong at the edge or ingress. Startup needs no external API calls.

`-dev` reloads templates and reviews from disk. Production embeds them in the binary.

Last.fm results are cached for 30 seconds. One request refreshes an expired
result while concurrent requests use the previous track. Upstream errors retain
the last successful result and back off for 30 seconds; a cold cache or the
request performing the refresh can still wait up to the three-second client
timeout. Disconnecting that request cancels its upstream fetch.

Run `go test -race ./...` and `go vet ./...` to check the Go code.
