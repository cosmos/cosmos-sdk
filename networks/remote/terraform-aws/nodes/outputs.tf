// The cluster name
output "name" {
  value = "${var.name}"
}

// The list of cluster instance IDs
output "instances" {
  value = ["${aws_instance.node.*.id}"]
}

// The list of cluster instance public IPs
output "public_ips" {
  value = ["${aws_instance.node.*.public_ip}"]
}

