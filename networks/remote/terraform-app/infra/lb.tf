resource "aws_lb" "lb" {
  name            = "${var.name}"
  subnets         = ["${aws_subnet.subnet.*.id}"]
#  security_groups = ["${split(",", var.lb_security_groups)}"]
  tags {
    Name    = "${var.name}"
  }
#  access_logs {
#    bucket = "${var.s3_bucket}"
#    prefix = "ELB-logs"
#  }
}

resource "aws_lb_listener" "lb_listener" {
  load_balancer_arn = "${aws_lb.lb.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.lb_target_group.arn}"
    type             = "forward"
  }
}

resource "aws_lb_listener_rule" "listener_rule" {
#  depends_on   = ["aws_lb_target_group.lb_target_group"]
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
  port     = "80"
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.vpc.id}"
  tags {
    name = "${var.name}"
  }
#  stickiness {
#    type            = "lb_cookie"
#    cookie_duration = 1800
#    enabled         = "true"
#  }
#  health_check {
#    healthy_threshold   = 3
#    unhealthy_threshold = 10
#    timeout             = 5
#    interval            = 10
#    path                = "${var.target_group_path}"
#    port                = "${var.target_group_port}"
#  }
}

