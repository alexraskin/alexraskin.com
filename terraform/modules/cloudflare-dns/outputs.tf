output "acm_record_names" {
  value = [for r in cloudflare_dns_record.acm : r.name]
}
