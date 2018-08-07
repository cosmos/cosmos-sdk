variable "name" {
  description = "The cluster name, e.g remotenet"
}

variable "regions" {
  description = "Regions to launch in"
  type = "list"
  default = ["AMS2", "TOR1", "LON1", "NYC3", "SFO2", "SGP1", "FRA1"]
}

variable "ssh_private_file" {
  description = "SSH private key filename to use to connect to the nodes"
  type = "string"
}

variable "ssh_public_file" {
  description = "SSH public key filename to copy to the nodes"
  type = "string"
}

variable "instance_size" {
  description = "The instance size to use"
  default = "2gb"
}

variable "servers" {
  description = "Desired instance count"
  default     = 4
}

