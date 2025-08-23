variable "vms_enabled" {
  default = false
}

variable "api_server_count" {
  type = number
  default = 1
}

variable "consumer_server_count" {
  type = number
  default = 0
}