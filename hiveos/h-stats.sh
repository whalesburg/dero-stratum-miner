#!/usr/bin/env bash

stats_raw=`curl http://localhost:44000/ -X POST -H "Content-Type: application/json" --silent -d '{"id":0,"jsonrpc":"2.0","method":"miner_getstat1"}' `
echo $stats_raw > /tmp/stats_raw

if [[ $? -ne 0 || -z $stats_raw ]]; then
    echo -e "${YELLOW}Failed to read $miner from localhost:44000${NOCOLOR}"
else
    local temp=`paste <(cat /sys/class/thermal/thermal_zone*/type) <(cat /sys/class/thermal/thermal_zone*/temp) | column -s $'\t' -t | sed 's/\(.\)..$/.\1°C/' | grep x86 | awk '{ print substr( $2, 0, length($2)-4 ) }' `
    local hs=`jq -c -r '.result[3]' <<< $stats_raw `
    local khs=`printf %.3f "$(($hs))e-3"`
    local hs_units="hs"
    local fan="[0]"
    local uptime=`jq -c -r '.result[1]' <<< $stats_raw`
    local ver=`jq -c -r '.result[0]' <<< $stats_raw`
    local shares=`jq -c -r '.result[2]' <<< $stats_raw`
    local acc=`echo $shares | awk -F";" '/1/ {print $2}'`
    local rej=`echo $shares | awk -F";" '/1/ {print $3}'`
    local bus_numbers=0
    local algo="astrobwt"
    stats=$(jq --argjson hs "$hs" --arg hs_units "$hs_units" --argjson temp "$temp" --argjson fan "$fan" --argjson uptime "$uptime" --arg ver "$ver" --arg acc "$acc" --arg rej "$rej" --argjson bus_numbers "$bus_numbers" --arg algo "AstroBWTv3" --arg total_khs "$khs" '{hs: [$hs] , $hs_units, temp:[$temp], $fan, $uptime, $ver, ar: [ $acc, $rej ], $bus_numbers, $algo, $total_khs}' <<< "$stats_raw")
fi

[[ -z $khs ]] && khs=0
[[ -z $stats ]] && stats="null"
