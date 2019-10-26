#!/bin/sh
set -e

if [ -n "$fyne_uid" -a "$fyne_uid" -gt 0 ]; then
    id -u $fyne_uid > /dev/null 2>&1 || adduser -q --disabled-password --gecos "" --uid $fyne_uid fyne 
    chown $fyne_uid /go
    eval exec gosu $fyne_uid $@
fi

eval exec $@
