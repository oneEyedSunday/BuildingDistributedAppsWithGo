# Building Distributed Applications with Go
Code workthrough / Educational project of the Pluralsight course ["Building Distributed Applications with Go"](https://www.pluralsight.com/courses/building-distributed-applications-go)

A walkthrough of building the basic elements of a distributed application architecture from first principles, using only Go’s standard library.


## Describe what is actually here...

Hub mixed etc


## Running app

docker build -t oneeyeedsunday_building .

docker run oneeyeedsunday_building ./registry_bin

docker-compose -p buildingDistributedAppsWithGo -f cluster/docker/docker-compose.yml up --build