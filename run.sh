#!/bin/bash
# proxy "go run cmd/worker/main.go" intercept and save packets using tcpdump
# name save file with "proxy-dump-<timestamp:formatted(YYYY-mm-dd-h:m:s)>".pcap
# run tcpdump with the following command:
# prxoy this app run and capture only smpp calls and responses (talk to smpp server throu it )
# incoming and outgoing packets on port 2775
tcpdump -i any -w proxy-dump-$(date +"%Y-%m-%d-%H:%M:%S").pcap -s 0 -nn -tttt -A 'port 2775'