#!/bin/bash
service tor start

pwd=$(tor --hash-password not_secure | tail -1)
echo HashedControlPassword $pwd >> /etc/tor/torrc
echo ControlPort 9051 >> /etc/tor/torrc

kill $(ps -ef | grep /usr/bin/tor | head -1 | awk {' print $2 '})
service tor restart

while true
do
    echo "Generating user"
    python3 ./tor.py --gender male
    sleep 15
done
