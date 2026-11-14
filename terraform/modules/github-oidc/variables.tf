variable "repo" {
  type        = string
  description = "owner/name; only its main branch may assume the role."
}

variable "role_name" {
  type = string
}

variable "state_bucket" {
  type = string
}

variable "state_key_prefix" {
  type = string
}

variable "site_bucket_arn" {
  type = string
}
