FROM alpine 

COPY kube-start-stop /kubey

ENTRYPOINT ["/kubey"]

