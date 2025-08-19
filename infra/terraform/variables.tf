variable "vms_enabled" {
  default = true
}

variable "api_server_count" {
  type = number
  default = 4
}

variable "consumer_server_count" {
  type = number
  default = 0
}