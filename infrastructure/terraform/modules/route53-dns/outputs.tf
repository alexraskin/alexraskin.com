output "acm_record_names" {
  value = [for r in aws_route53_record.acm : r.fqdn]
}
