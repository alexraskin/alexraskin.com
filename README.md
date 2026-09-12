# alexraskin.com

My personal website — [alexraskin.com](https://alexraskin.com).

Built with Go, HTML templates, and plain CSS. Includes my latest Last.fm track
and a collection of Franzbrötchen reviews. Templates and static assets are
embedded in a single binary; review photos are served from Cloudflare R2.

Run locally with live template reloading:

```sh
mise run dev
```

Deployment manifests live in my [infrastructure repo](https://github.com/alexraskin/infrastructure/tree/main/apps/base/alexraskin-com).
