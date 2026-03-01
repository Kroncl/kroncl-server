#!/bin/bash
# Ожидание доступности хоста и порта
# Использование: ./wait-for-it.sh host:port [-t timeout] [-- command args]

set -e

hostport="$1"
shift
cmd=()

timeout=30

while [[ $# -gt 0 ]]; do
    case "$1" in
        -t)
            timeout="$2"
            shift 2
            ;;
        --)
            shift
            cmd=("$@")
            break
            ;;
        *)
            break
            ;;
    esac
done

# Парсим host:port
host=$(echo "$hostport" | cut -d: -f1)
port=$(echo "$hostport" | cut -d: -f2)

echo "Waiting for $host:$port..."

start_ts=$(date +%s)
while true; do
    if nc -z "$host" "$port" 2>/dev/null; then
        break
    fi
    current_ts=$(date +%s)
    elapsed=$((current_ts - start_ts))
    if [[ $elapsed -ge $timeout ]]; then
        echo "Timeout: $host:$port not available after $timeout seconds"
        exit 1
    fi
    sleep 1
done

echo "$host:$port is available"

if [[ ${#cmd[@]} -gt 0 ]]; then
    exec "${cmd[@]}"
fi