variable "vms_enabled" {
  default = false
}

variable "api_server_count" {
  type = number
  default = 3
}

variable "consumer_server_count" {
  type = number
  default = 0
}