# Load balancer and related config for Bacalhau public endpoint
resource "aws_lb" "labdao_public" {
  name               = "labdao-${var.environment}-public"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.labdao_public_bacalhau.id, aws_security_group.labdao_public_plex.id, aws_security_group.labdao_egress_all.id]
  ip_address_type    = "ipv4"
  subnets            = data.aws_subnets.default_filtered.ids
}

resource "aws_lb_listener" "labdao_public_1234" {
  load_balancer_arn = aws_lb.labdao_public.arn
  port              = "1234"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.labdao_requester_bacalhau_tg.arn
    type             = "forward"
  }
}

resource "aws_lb_listener" "labdao_public_5001" {
  load_balancer_arn = aws_lb.labdao_public.arn
  port              = "5001"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.labdao_requester_ipfs_tg.arn
    type             = "forward"
  }
}

resource "aws_lb_target_group" "labdao_requester_bacalhau_tg" {
  name     = "labdao-${var.environment}-requester-tg"
  port     = 1234
  protocol = "HTTP"

  # PUT in default VPC for now
  vpc_id = data.aws_vpc.default.id

  # deregistration_delay = 5

  # NOTE: amount time for targets to warm up before the load balancer sends them a full share of requests
  slow_start = 60

  health_check {
    interval            = 5
    path                = "/readyz"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 2
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

resource "aws_lb_target_group" "labdao_requester_ipfs_tg" {
  name     = "labdao-${var.environment}-requester-ipfs-tg"
  port     = 5001
  protocol = "HTTP"

  # PUT in default VPC for now
  vpc_id = data.aws_vpc.default.id

  # deregistration_delay = 5

  # slow_start = 60

  # TODO: need to figure out healthcheck for IPFS
  # health_check {
  #   interval            = 5
  #   port                = "traffic-port"
  #   protocol            = "TCP"
  #   timeout             = 2
  #   healthy_threshold   = 3
  #   unhealthy_threshold = 3
  # }
}

# public dns record for recepter
resource "cloudflare_record" "labdao_requester" {
  zone_id = var.cloudflare_zone_id
  name    = "bacalhau.${var.environment}"
  value   = aws_lb.labdao_public.dns_name
  type    = "CNAME"
  ttl     = 3600
}
