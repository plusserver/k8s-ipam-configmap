FROM golang:latest AS build
WORKDIR /go/src/app
COPY . .
RUN go get k8s.io/client-go/...
RUN go get -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-ipam-configmap *.go

FROM scratch
COPY --from=build /go/src/app/k8s-ipam-configmap /k8s-ipam-configmap
CMD ["/k8s-ipam-configmap"]
