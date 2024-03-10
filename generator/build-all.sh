for description in input/*;
do 
    name=$(awk -F'[/.]' '{print $2}' <<<$description)
    # sometimes the base image will be deleted but I have no idea..
    # build base image
    sudo docker build -t "$(hostname -f)/hydragen-base" ..

    # generate yamls and tag/push image
    image=$(sudo ./generator.sh preset input/$name.json | tail -1 | awk -F':' '{print $2":"$3}')
    echo $image
    sudo docker tag $image yizhengx/hydragen:$name
    sudo docker push yizhengx/hydragen:$name
    sudo docker rmi yizhengx/hydragen:$name # untag image

    # modify yaml files to change image
    sudo rm -rf $name-yamls
    sudo mkdir $name-yamls
    sudo chmod 777 k8s/*
    sudo mv k8s/* $name-yamls/
    for file in $name-yamls/*;
    do 
        sudo sed -i "/image:/c\                  image: yizhengx/hydragen:${name}" $file
        sudo sed -i "/imagePullPolicy:/c\                  imagePullPolicy: IfNotPresent" $file
    done
done