sudo sysctl -w net.core.somaxconn=8192

sudo sysctl -w net.ipv4.tcp_max_syn_backlog=8192

sudo sysctl -w net.core.netdev_max_backlog=16384

sudo sysctl -w net.ipv4.tcp_fin_timeout=30

sudo sysctl -w net.ipv4.tcp_tw_reuse=1

ulimit -n 1048576
