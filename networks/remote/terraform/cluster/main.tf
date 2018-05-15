resource "digitalocean_tag" "cluster" {
  name = "${var.name}"
}

resource "digitalocean_ssh_key" "cluster" {
  name       = "${var.name}"
  public_key = "${file(var.ssh_key)}"
}

resource "digitalocean_droplet" "cluster" {
  name = "${var.name}-node${count.index}"
  image = "centos-7-x64"
  size = "${var.instance_size}"
  region = "${element(var.regions, count.index)}"
  ssh_keys = ["${digitalocean_ssh_key.cluster.id}"]
  count = "${var.servers}"
  tags = ["${digitalocean_tag.cluster.id}"]

  lifecycle = {
	prevent_destroy = false
  }

  connection {
    timeout = "30s"
  }

}

