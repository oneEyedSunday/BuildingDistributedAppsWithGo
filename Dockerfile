# build app
FROM golang:1.22.1 AS main-env
RUN mkdir /app
ADD src /app/
ADD src/cmd/ /app/cmd/
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"  -o log_bin ./cmd/runLogService.go
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"  -o grading_bin ./cmd/runGradingService.go
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"  -o registry_bin ./cmd/runRegistryService.go
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"  -o teacher_bin ./cmd/runTeacherPortal.go

# run artifacts
# FROM scratch
FROM alpine:3.14
# RUN apk add --no-cache ls
WORKDIR /app
COPY --from=main-env /app/log_bin /app
COPY --from=main-env /app/grading_bin /app
COPY --from=main-env /app/registry_bin /app
COPY --from=main-env /app/teacher_bin /app
EXPOSE 3000
EXPOSE 4000
EXPOSE 6000
EXPOSE 5000

# TODO make this a dynamic listing
CMD ["echo", "specify a service; options are registry_bin, log_bin, grading_bin, teacher_bin"]