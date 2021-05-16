import subprocess
import logging
import time
import datetime

def ensure_ready_probe():
    # check logs
    ready = False
    attempts = 0

    while not ready:
        if attempts > 10:
            break
        r = {}
        try:
            r = open('./logger.log', 'r')
        except FileNotFoundError:
            print('File not found')
            attempts += 1
            time.sleep(5)
            continue

        lines = r.readlines()
        last = lines[-5:]
        for line in last:
            if 'miner' in line and 'speed' in line:
                ready = True

        print('Logging not ready')
        r.close()
        attempts += 1
        time.sleep(5)
        
    # readiness probe
    if ready:
        f = open("/tmp/ready", "a")
        f.close()
    return ready

# ready = ensure_ready_probe(misconfigs)
pool = "--url=pool.xmrpool.eu:3333"
# Wikileaks
wallet = "--user=453VWT5GEkXGc2J9asRpXpRkjoCGKCJr96rndm2VMe5yECiAcUB3h8pFxZ8YGbmbGmVefwWHPXmLR69Vw1sVNWz5TsFqYbK"
logfile = "./logger.log"

logging.info("Starting TOR")
tor = subprocess.call(['service', 'tor', 'start'])
logging.info("Starting to mine")
miner = subprocess.Popen(['proxychains4', '/xmrig/build/xmrig', f'{pool}', '--donate-level=3', f'{wallet}', '--pass=docker', '-k', '--coin=monero', f'--log-file={logfile}', '-B', '--print-time=5'])

ready = ensure_ready_probe()
while ready:
    print("Cryptomining behaviour started in cluster")
    time.sleep(30)
else:
    print("Cryptomining behaviour could not be started")