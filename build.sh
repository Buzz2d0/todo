#!/bin/bash

#wailsPack
function wailsPack() {
    # pack wails app
    wails build -p
    # remove 
    mac_info_plist="./info.plist"
    if [ -f "$mac_info_plist" ]; then
        rm $mac_info_plist
    fi
}

wailsPack
