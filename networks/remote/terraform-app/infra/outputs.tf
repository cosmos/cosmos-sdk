// The cluster name
output "name" {
  value = "${var.name}"
}

// The list of cluster instance IDs
output "instances" {
  value = ["${aws_instance.node.*.id}"]
}

#output "instances_count" {
#  value = "${length(aws_instance.node.*)}"
#}

// The list of cluster instance public IPs
output "public_ips" {
  value = ["${aws_instance.node.*.public_ip}"]
}

// Name of the ALB
output "lb_name" {
  value = "${aws_lb.lb.dns_name}"
}

