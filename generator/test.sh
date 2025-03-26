
for file in chain-d2-grpc-async/yamls/*;
do 
    # file is x.yaml, x is the name of the service, extract x in uppercase
    service=$(echo $file | awk -F'/' '{print $NF}' | awk -F'.' '{print toupper($1)}')
    sudo sed -i "s/9999/\"\${PROCESSING_TIME_$service}\"/g" $file
done