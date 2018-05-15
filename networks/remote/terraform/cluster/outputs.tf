// The cluster name
output "name" {
  value = "${var.name}"
}

// The list of cluster instance IDs
output "instances" {
  value = ["${digitalocean_droplet.cluster.*.id}"]
}

// The list of cluster instance public IPs
output "public_ips" {
  value = ["${digitalocean_droplet.cluster.*.ipv4_address}"]
}

