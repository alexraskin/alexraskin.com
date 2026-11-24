# website

Hugo site ([terminal](https://github.com/panr/hugo-theme-terminal) theme, Catppuccin Macchiato colors in `assets/css/terminal.css`) served from S3 + CloudFront. `mise run build` renders into `public/` (gitignored).

```sh
git submodule update --init  # theme
mise trust && mise install && pre-commit install  # first time only
mise run dev                 # hugo server
```

## Infrastructure and deployment

Shared with the other Hugo sites through
[alexraskin/terraform-hugo-module](https://github.com/alexraskin/terraform-hugo-module):
its `site` and `route53_dns` modules make up `infrastructure/terraform/`, its
`deploy_role` the CI role in `bootstrap/` (applied by hand), and its reusable
workflow runs CI. `mise run deploy` (`hugo deploy`) uploads `public/` with the
targets and cache rules under `[deployment]` in `hugo.toml`.

```sh
mise run lint                    # pre-commit hooks
mise run tf:init && mise run tf:plan && mise run tf:apply
mise run build && mise run deploy
```

`.github/workflows/deploy.yml` lints and builds every push/PR and, on `main`,
applies Terraform and deploys via GitHub OIDC (role `github-deploy-alexraskin-prod`,
`prod` environment only).
