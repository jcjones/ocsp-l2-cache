#!/usr/bin/env python3

import iterm2
import netifaces as ni

my_ip = ni.ifaddresses("en0")[ni.AF_INET][0]["addr"]


async def main(connection):
    app = await iterm2.async_get_app(connection)
    window = app.current_window
    if window is not None:
        t = await window.async_create_tab(
            command="/usr/local/bin/docker run -p 6379:7000 redis:4 --port 7000"
        )
        await t.async_set_title("Redis")
    else:
        print("No current window")


iterm2.run_until_complete(main)
