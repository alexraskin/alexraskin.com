variable "github_repository" {
  description = "GitHub repository allowed to assume the deploy roles, as owner/name."
  type        = string
  default     = "alexraskin/alexraskin.com"
}

variable "github_sub_prefix" {
  description = "OIDC sub claim prefix for the repo; the trust appends :environment:<env>."
  type        = string
  default     = "repo:alexraskin/alexraskin.com"
}

variable "environments" {
  description = "One deploy role per env; each trusts only the GitHub environment of the same name."
  type        = set(string)
  default     = ["prod"]
}
