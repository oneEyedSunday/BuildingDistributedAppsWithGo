# Building Distributed Applications with Go
Code workthrough / Educational project of the Pluralsight course ["Building Distributed Applications with Go"](https://www.pluralsight.com/courses/building-distributed-applications-go)

A walkthrough of building the basic elements of a distributed application architecture from first principles, using only Go’s standard library.


## Describe what is actually here...

Hub mixed etc (TODO describe what architecture we are building on)


## Running app

with docker, i've decided to build the multiple service artifacts with one Dockerfile, the image itself having multiple binaries.

The services themselves can be run via  `docker run <image name> ./<bin_name>`

Its a bit involved to run on docker due to the networking involved, but ive added the following snippet which builds the image and runs the container 

```sh
docker build -t oneeyeedsunday_building .
docker network create school_app_network
docker run -d --name school_app_registry -e REGISTRY_SERVICE_HOST=school_app_registry oneeyeedsunday_building ./registry_bin
docker network connect school_app_network school_app_registry
docker run -d --name school_app_log --network=school_app_network -e REGISTRY_SERVICE_HOST=school_app_registry \
-e SERVICE_HOSTNAME='school_app_log' \
oneeyeedsunday_building ./log_bin
docker run -d --name school_app_grading --network=school_app_network --publish 6600:6000 \
-e REGISTRY_SERVICE_HOST=school_app_registry \
-e SERVICE_HOSTNAME='school_app_grading' \
oneeyeedsunday_building ./grading_bin
```

You can reach the grading service from your local computer at
```sh
curl http://localhost:6600/students
```

run all services with docker compose

```sh
docker-compose -p buildingDistributedAppsWithGo -f cluster/docker/docker-compose.yml up --build
```