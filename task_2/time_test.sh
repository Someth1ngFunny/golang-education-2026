#!/usr/bin/env bash
TIMEFORMAT="%R s"
for w in 1 10; do
    echo -n "$w workers: "
    time ./main.exe --from 10 --to 100 --workers "$w" >/dev/null 2>&1
done