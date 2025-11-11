data "vms" "this" {
  name = "instance01-test"
  fail_on_empty = true
}

output "attachment_set" {
  value = data.vms.this
}

# {
#   "fail_on_empty" = true
#   "id" = "instance01-test"
#   "name" = "instance01-test"
#   "vms" = toset([
#     {
#       "id" = "f5742dc5-9443-4a0d-b741-100b9b09b4b9"
#       "ips" = tolist([
#         "10.10.10.10",
#         "fe80::546f:a0ff:feba:ff",
#         "10.10.10.11",
#         "fe80::546f:a0ff:feba:100",
#       ])
#       "status" = "up"
#     },
#   ])
# }