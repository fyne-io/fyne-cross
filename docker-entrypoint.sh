#!/bin/sh
set -e

if [ -z "$fyne_uid" ] || [ $fyne_uid -eq 0 2>/dev/null ]; then
    eval exec $@    
    exit $?
fi

if [ $fyne_uid -gt 0 2>/dev/null ]; then
    id -u $fyne_uid > /dev/null 2>&1 || adduser -q --disabled-password --gecos "" --uid $fyne_uid fyne 
    chown $fyne_uid /go
    eval exec gosu $fyne_uid $@        
    exit $?
fi

echo "fyne_uid must be >= 0"
exit 1
