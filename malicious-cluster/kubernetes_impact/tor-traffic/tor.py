#!/usr/bin/env python3
from stem import Signal
from stem.control import Controller
from argparse import ArgumentParser as AP
from fake_useragent import UserAgent
from requests import get
from os import environ
from bs4 import BeautifulSoup as bs
from colored import fg, attr
from re import compile, findall, IGNORECASE
from prettytable import PrettyTable as PT
from random import randint


row_data_cell_re = compile(r'<td [\w="]+>([\w\s\.@:()\'-0-9]+)<\/td>')


def extract_row_data_cells(row: str):
        return row_data_cell_re.findall(row, IGNORECASE)

def ensure_ready_probe():
    f = open("/tmp/ready", "a")
    f.close()


def make_request(url: str, gender: str):
    tor_proxy = {
            "http": "socks5h://127.0.0.1:9050",
            "https": "socks5h://127.0.0.1:9050"
        }

    random_user_agent = UserAgent().random
    headers = {
        "User-Agent": random_user_agent
        }

    if gender == "female":
        url = url + "?gender=f"
    elif gender == "male":
        url = url + "?gender=m"
    else: 
        url = url + "?gender=r"

    return get(url, headers=headers, proxies=tor_proxy)


def create_new_tor_ip():
    with Controller.from_port(port=9051) as controller:
        controller.authenticate(password=environ["TOR_PASS"])
        controller.signal(Signal.NEWNYM)

    
if __name__ == "__main__":
    hidden_service_url = "http://elfq2qefxx6dv3vy.onion/fakeid.php"

    parser = AP(description="Generate identity to browse the Dark Web")
    parser.add_argument(
        "-g", "--gender",
        choices=("female", "male", "random"),
        default="random",
        help="Select the gender to generate a Fake Id for"
        )
    args = parser.parse_args()

    resp = make_request(hidden_service_url, args.gender)
    if resp.status_code == 200:
        ensure_ready_probe()
        print("Tor traffic in pod successful")