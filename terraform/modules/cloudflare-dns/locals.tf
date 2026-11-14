locals {
  acm = { for o in var.acm_validation_options : o.domain_name => o }
}
