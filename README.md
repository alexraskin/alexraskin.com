# website

Hugo site ([terminal](https://github.com/panr/hugo-theme-terminal) theme, Catppuccin Macchiato colors in `assets/css/terminal.css`) served from S3 + CloudFront. `mise run build` renders into `public/` (gitignored).

```sh
git submodule update --init  # theme
mise trust                   # first time only
mise run dev                 # hugo server
```

## Terraform

`infrastructure/terraform/` mirrors `public/` to S3 + CloudFront. `tf:plan` runs `build` first.

```sh
mise run tf:plan
mise run tf:apply
```

- CI (`.github/workflows/deploy.yml`) runs `terraform apply` on push to `main` via GitHub OIDC (role `github-deploy-website`, main branch only).
- CI role has read-only IAM, so changes to `modules/github-oidc` must be applied locally.

| Module | Does |
|---|---|
| `s3-site` | Private bucket, uploads `public/` with MIME types + cache headers |
| `cloudfront` | Distribution with OAC; function maps `/dir/` -> `/dir/index.html`, 301s `/dir` -> `/dir/` |
| `route53-dns` | Alias records + ACM validation records |
| `github-oidc` | OIDC provider + deploy role for CI |
