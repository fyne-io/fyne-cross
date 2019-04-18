#!/bin/sh
set -e

if [ -n "$fyne_uid" ]; then
    adduser -q --disabled-password --gecos "" --uid $fyne_uid fyne 
    chown $fyne_uid /go
    exec gosu $fyne_uid $@
fi

exec $@
