cd $(dirname $0)

for description in input/large-scale.json;
do 
    name=$(awk -F'[/.]' '{print $2}' <<<$description)
    # sometimes the base image will be deleted but I have no idea..
    # build base image
    sudo docker build -t "$(hostname -f)/hydragen-base" ..

    # generate yamls and tag/push image
    image=$(sudo bash generator.sh preset input/$name.json | tail -1 | awk -F':' '{print $2":"$3}')
    echo $image
    sudo docker tag $image yizhengx/hydragen:$name
    sudo docker push yizhengx/hydragen:$name
    sudo docker rmi yizhengx/hydragen:$name # untag image

    # modify yaml files to change image
    sudo rm -rf $name/yamls
    sudo mkdir -p $name/yamls
    sudo chmod 777 k8s/*
    sudo mv k8s/* $name/yamls/
    for file in $name/yamls/*;
    do 
        # file is x.yaml, x is the name of the service, extract x in uppercase
        service=$(echo $file | awk -F'/' '{print $NF}' | awk -F'.' '{print toupper($1)}')
        sudo sed -i "s/9999/\${PROCESSING_TIME_$service}/g" $file
        sudo sed -i 's/\${SLOWPOKE_DELAY_MICROS_[^}]*}/"&"/g' $file
        sudo sed -i 's/\${SLOWPOKE_POKER_BATCH_THRESHOLD_[^}]*}/"&"/g' $file
        sudo sed -i 's/\${SLOWPOKE_IS_TARGET_SERVICE_[^}]*}/"&"/g' $file
        sudo sed -i 's/${SLOWPOKE_PRERUN}/"${SLOWPOKE_PRERUN}"/g' $file
        # replace 9999 with string "${PROCESSING_TIME_$service}"
        sudo sed -i "/image:/c\                  image: yizhengx/hydragen:${name}" $file
        sudo sed -i "/imagePullPolicy:/c\                  imagePullPolicy: Always" $file
        sudo sed -i 's|yizhengx/hydragen:large-scale|yizhengx/mucache:synthetic-pokerpp-0maxconn|g' $file
    done
done
