# http-grpc-websocket

A demo project about how you can run HTTP, gRPC and websocket in Cloud Run using
a single port. You can follow this tutorial [HTTP, gRPC, and websocket on Google Cloud Run](https://hodo.dev/posts/post-41-gcp-cloudrun-grpc-http-ws/) to find out more.

## Run and test locally

Run 

```bash
go run main.go
```

On a separate terminal,

test HTTP

```bash
curl http://localhost:8080/ping
```

test gRPC

```bash
go run clients/grpc/main.go
```

test Websocket

```bash
go run clients/ws/main.go
```

## Run and test on Cloud Run with authentication 

Set your variables 

```bash
PROJECT_ID=$(gcloud config list project --format='value(core.project)')
REGION=us-central1
```

Create a service account

```bash
gcloud iam service-accounts create demo-grpc-sa \
  --display-name "Demo service account for gRPC on Cloud Run"
```

Create a new key for this service account

```bash
gcloud iam service-accounts keys create key.json \
  --iam-account demo-grpc-sa@${PROJECT_ID}.iam.gserviceaccount.com
```

Build the image

```bash
gcloud builds submit --tag gcr.io/$PROJECT_ID/grpcwebapp --project $PROJECT_ID .
```

Deploy the service 

```bash
gcloud run deploy grpcwebapp \
--image gcr.io/$PROJECT_ID/grpcwebapp \
--set-env-vars=AUTH_SERVICE_ACCOUNTS="demo-grpc-sa@${PROJECT_ID}.iam.gserviceaccount.com",AUTH_AUDIENCE=webapp \
--allow-unauthenticated \
--timeout=10m \
--project $PROJECT_ID \
--region $REGION
```

Get the public URL

```bash
SERVICE_URL=$(gcloud run services describe grpcwebapp --platform managed --region $REGION --format 'value(status.url)')
echo $SERVICE_URL
SERVICE_HOST=$(echo "$SERVICE_URL" | awk -F/ '{print $3}')
echo $SERVICE_HOST
```

test HTTPS

```bash
curl -v $SERVICE_URL/ping
```

test gRPC

```bash
go run clients/grpc/main.go --server "$SERVICE_HOST:443" --key key.json --insecure=false
```

test Websocket

```bash
go run clients/ws/main.go --server "wss://$SERVICE_HOST/ws" --key key.json
```

## Generate protobuffer files.

```bash
protoc \
--proto_path=grpc/proto \
--go_out=plugins=grpc:. \
./grpc/proto/*.proto
```

## References

This example was based on this post [Serving gRPC+HTTP/2 from the same Cloud Run container](https://ahmet.im/blog/grpc-http-mux-go/) 
written by Ahmet Alp Balkan (https://github.com/ahmetb).
