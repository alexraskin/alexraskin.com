# website

Static site served from S3 + CloudFront. Everything served lives in `public/`.

```sh
mise trust        # first time only
mise run dev      # local dev server
```

## Terraform (AWS mirror)

`terraform/` mirrors `public/` to S3 + CloudFront, plus a $5/mo account-wide AWS budget alert. State lives in S3 (`alexraskin-tfstate`, key `website/terraform.tfstate`, native lockfile).

```sh
mise run tf:plan
mise run tf:apply
```

Needs AWS creds, `CLOUDFLARE_API_TOKEN` (Zone:DNS:Edit + Zone:Read), and `budget_email` via `terraform/terraform.tfvars` or `TF_VAR_budget_email` (gitignored; repo is public).

- `attach_domain` (default `false`): site only on `*.cloudfront.net`. Set `true` to add ACM cert + Cloudflare DNS for the apex and `www`.
- CI (`.github/workflows/deploy.yml`) runs `terraform apply` on push to `main` via GitHub OIDC (role `github-deploy-website`, main branch only).
- CI role has read-only IAM, so changes to `modules/github-oidc` must be applied locally.

| Module | Does |
|---|---|
| `s3-site` | Private bucket, uploads `public/` with MIME types + cache headers |
| `cloudfront` | Distribution with OAC, pretty-URL rewrite function |
| `cloudflare-dns` | CNAMEs + ACM validation records |
| `github-oidc` | OIDC provider + deploy role for CI |
