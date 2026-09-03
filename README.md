# alexraskin.com

Source for my personal website.

```bash
mise run dev      # local, live template reload
mise run docker-up # docker
```

## Adding a Franzbrötchen review

1. `mise run add-review <photo.jpg> [stem]` — writes AVIF and JPEG copies at the
   widths the page serves into `assets/images/franzbroetchen/`, and prints the
   JSON entry. Only these copies are committed; keep the original elsewhere.
2. Append the printed object to `data/franzbroetchen.json` and fill it in:

```json
{
  "place": "Bäckerei Name",
  "location": "Hamburg, Germany",
  "date": "2025-09-01",
  "rating": 4,
  "photo": "/assets/images/franzbroetchen/photo.jpg",
  "note": "Optional.",
  "url": "https://optional-bakery-link"
}
```

`rating` is 1-5, `date` is `YYYY-MM-DD`, and entries render newest first. The
`photo` path is a logical name: the page picks up whatever `<stem>-<width>.avif`
and `<stem>-<width>.jpg` files sit beside it, so encoding another size later
needs no code change. Image dimensions are read from the file. A malformed entry
fails the build at startup rather than rendering a broken page.
