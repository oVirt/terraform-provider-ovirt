variable "storagedomain_id" {
  type = string
  description = "ID of the storage domain to create the disk on."
}

variable "cluster_id" {
  type = string
}

variable "username" {
  type = string
}
variable "password" {
  type = string
}
variable "url" {
  type = string
}
variable "tls_ca_files" {
  type    = list(string)
  default = []
}
variable "tls_ca_dirs" {
  type    = list(string)
  default = []
}
variable "tls_insecure" {
  type    = bool
  default = false
}
variable "tls_ca_bundle" {
  type    = string
  default = ""
}
variable "tls_system" {
  type        = bool
  default     = true
  description = "Take TLS CA certificates from system root. Does not work on Windows."
}
variable "mock" {
  type    = bool
  default = true
}