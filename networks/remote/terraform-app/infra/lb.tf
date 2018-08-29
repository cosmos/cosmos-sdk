resource "aws_lb" "lb" {
  name            = "${var.name}"
  subnets         = ["${aws_subnet.subnet.*.id}"]
  security_groups = ["${aws_security_group.secgroup.id}"]
  tags {
    Name    = "${var.name}"
  }
#  access_logs {
#    bucket = "${var.s3_bucket}"
#    prefix = "lblogs"
#  }
}

resource "aws_lb_listener" "lb_listener" {
  load_balancer_arn = "${aws_lb.lb.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-Ext-2018-06"
  certificate_arn   = "${var.certificate_arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.lb_target_group.arn}"
    type             = "forward"
  }
}

resource "aws_lb_listener_rule" "listener_rule" {
  listener_arn = "${aws_lb_listener.lb_listener.arn}"
  priority     = "100"
  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.lb_target_group.id}"
  }
  condition {
    field  = "path-pattern"
    values = ["/"]
  }
}

resource "aws_lb_target_group" "lb_target_group" {
  name     = "${var.name}"
  port     = "26657"
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.vpc.id}"
  tags {
    name = "${var.name}"
  }
  health_check {
    path                = "/health"
  }
}

