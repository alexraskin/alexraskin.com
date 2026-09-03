# alexraskin.com

Source for my personal website.

```bash
mise run dev      # local, live template reload
mise run docker-up # docker
```

## Adding a Franzbrötchen review

1. Drop the photo in `assets/images/franzbroetchen/`.
2. Append an object to `data/franzbroetchen.json`:

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

`rating` is 1-5, `date` is `YYYY-MM-DD`, and entries render newest first.
Image dimensions are read from the file, so nothing else needs updating. A
malformed entry fails the build at startup rather than rendering a broken page.
