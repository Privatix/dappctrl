#!/usr/bin/env bash

prettify_json() {
    # Call "python -m json.tool <filename>" to prettify json and orders keys.
    cp "$1" "$1-buff"
    python -m json.tool "$1-buff" | cat > "$1"
    rm "$1-buff"
}

prettify_json dappctrl.config.json
prettify_json dappctrl-dev.config.json
prettify_json dappctrl-test.config.json