variable "vms_enabled" {
  default = false
}

variable "api_server_count" {
  type = number
  default = 2
}

variable "consumer_server_count" {
  type = number
  default = 0
}