# # for each file x.json, rename to x-sync.json
# for file in *.json; do
#     mv "$file" "${file%.json}-sync.json"
# done


# # for each x-sync.json, create a new file x-async.json, change "synchronous" to "asynchronous" in the file
# for file in *-sync.json; do
#     new_file="${file%-sync.json}-async.json"
#     cp "$file" "$new_file"
#     sed -i 's/synchronous/asynchronous/g' "$new_file"
# done

# for every file, remove 
# ,
#             "annotations": [
#               {
#                 "name": "sidecar.istio.io/statsInclusionPrefixes",
#                 "value": "cluster.outbound,cluster_manager,listener_manager,http_mixer_filter,tcp_mixer_filter,server,cluster.xds-grp,listener,connection_manager"
#               },
#               {
#                 "name": "sidecar.istio.io/statsInclusionRegexps",
#                 "value": "http.*"
#               },
#               {
#                 "name": "sidecar.istio.io/userVolume",
#                 "value": "[{\"name\":\"rate-limit-filter\",\"configMap\":{\"name\":\"rate-limit-filter\"}}]"
#               },
#               {
#                 "name": "sidecar.istio.io/userVolumeMount",
#                 "value": "[{\"mountPath\":\"/var/local/wasm\",\"name\":\"rate-limit-filter\"}]"
#               },
#               {
#                 "name": "sidecar.istio.io/proxyCPU",
#                 "value": "2000m"
#               },
#               {
#                 "name": "sidecar.istio.io/proxyCPULimit",
#                 "value": "2000m"
#               }
#             ]

# for file in *.json; do
#     echo "Processing $file"
#     jq '(.services[].clusters |= map(del(.annotations)))' $file > dump.json
#     mv dump.json $file
# done

for file in *.json; do
    echo "Processing $file"
    # change logging from "logging": true to "logging": false
    sed -i 's/"logging": true/"logging": false/g' $file
done